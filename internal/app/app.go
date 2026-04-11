package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"quick_app_switcher/internal/events"
	"quick_app_switcher/internal/input"
	"quick_app_switcher/internal/mru"
	"quick_app_switcher/internal/session"
	"quick_app_switcher/internal/ui"
	"quick_app_switcher/internal/win32"
	"quick_app_switcher/internal/windows"

	coderuntime "quick_app_switcher/internal/runtime"
)

const (
	classController = "QuickAppSwitcher.Controller"
	classOverlay    = "QuickAppSwitcher.Overlay"

	msgHookTabPressed    = win32.WM_APP + 1
	msgHookAltReleased   = win32.WM_APP + 2
	msgHookCancel        = win32.WM_APP + 3
	msgForegroundChanged = win32.WM_APP + 4
	msgTray              = win32.WM_APP + 5
	msgShutdownRequested = win32.WM_APP + 6
	msgThumbnailsReady   = win32.WM_APP + 7
	msgIconsReady        = win32.WM_APP + 8

	wmUser        = 0x0400
	ninSelect     = wmUser + 0
	ninKeySelect  = wmUser + 1
	wmContextMenu = 0x007B
)

type App struct {
	logger         *log.Logger
	controllerHwnd win32.HWND
	overlayHwnd    win32.HWND
	taskbarMsg     uint32

	hook         *input.Hook
	watcher      *events.ForegroundWatcher
	tray         *ui.Tray
	overlay      *ui.Overlay
	desktop      *windows.DesktopManager
	inventory    *windows.Inventory
	icons        *windows.IconCache
	thumbnails   *windows.ThumbnailCache
	mru          *mru.Store
	session      session.SwitchSession
	lastSnapshot windows.InventorySnapshot

	activateTarget      func(windows.WindowID) error
	isValidSwitchTarget func(windows.WindowID) bool

	windowProc       uintptr
	overlayProc      uintptr
	shuttingDown     atomic.Bool
	thumbnailWarmWG  sync.WaitGroup
	releaseModifiers func() error
}

func Run(logger *log.Logger) error {
	instance, err := coderuntime.AcquireSingleInstance()
	if err != nil {
		return err
	}
	defer func() {
		_ = instance.Release()
	}()

	unlock := coderuntime.LockOSThread()
	defer unlock()

	if err := win32.CoInitialize(); err != nil {
		return fmt.Errorf("initialize COM: %w", err)
	}
	defer win32.CoUninitialize()

	a := &App{
		logger:     logger,
		tray:       ui.NewTray(msgTray, "Quick App Switcher"),
		icons:      windows.NewIconCache(),
		thumbnails: windows.NewThumbnailCache(),
		mru:        mru.New(),
	}
	if err := a.initWindows(); err != nil {
		a.shutdown()
		return err
	}
	defer a.shutdown()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		a.requestShutdown()
	}()

	return a.loop()
}

func (a *App) initWindows() error {
	instance, err := win32.GetModuleHandle()
	if err != nil {
		return fmt.Errorf("get module handle: %w", err)
	}

	a.windowProc = syscall.NewCallback(a.controllerWndProc)
	a.overlayProc = syscall.NewCallback(a.overlayWndProc)
	if _, err := win32.RegisterWindowClass(classController, a.windowProc, instance, win32.LoadDefaultApplicationIcon()); err != nil {
		return fmt.Errorf("register controller class: %w", err)
	}
	if _, err := win32.RegisterWindowClass(classOverlay, a.overlayProc, instance, 0); err != nil {
		return fmt.Errorf("register overlay class: %w", err)
	}
	a.controllerHwnd, err = win32.CreateWindow(0, 0, classController, "Quick App Switcher", instance, uintptr(unsafe.Pointer(a)))
	if err != nil {
		return fmt.Errorf("create controller window: %w", err)
	}
	a.overlayHwnd, err = win32.CreateWindow(
		win32.WS_EX_TOPMOST|win32.WS_EX_TOOLWINDOW|win32.WS_EX_LAYERED|win32.WS_EX_NOACTIVATE,
		win32.WS_POPUP,
		classOverlay,
		"Quick App Switcher Overlay",
		instance,
		uintptr(unsafe.Pointer(a)),
	)
	if err != nil {
		return fmt.Errorf("create overlay window: %w", err)
	}
	if err := win32.SetLayeredWindowAlpha(a.overlayHwnd, 255); err != nil {
		return fmt.Errorf("initialize overlay alpha: %w", err)
	}
	a.overlay = ui.NewOverlay(a.overlayHwnd)
	a.taskbarMsg = ui.RegisterTaskbarCreated()

	a.desktop, _ = windows.NewDesktopManager()
	a.inventory = windows.NewInventory([]win32.HWND{a.controllerHwnd, a.overlayHwnd}, a.desktop)
	if err := a.refreshSnapshot(); err != nil {
		return fmt.Errorf("initial snapshot: %w", err)
	}
	current := windows.WindowID(win32.GetForegroundWindow())
	seed := make([]windows.WindowID, 0, len(a.lastSnapshot.Order)+1)
	if current != 0 {
		seed = append(seed, current)
	}
	seed = append(seed, a.lastSnapshot.Order...)
	a.mru.Seed(seed)

	a.watcher, err = events.NewForegroundWatcher(a.controllerHwnd, msgForegroundChanged)
	if err != nil {
		return fmt.Errorf("start foreground watcher: %w", err)
	}
	a.hook = input.New(a.controllerHwnd, msgHookTabPressed, msgHookAltReleased, msgHookCancel)
	if err := a.hook.Start(); err != nil {
		return fmt.Errorf("start keyboard hook: %w", err)
	}
	if err := a.tray.Add(a.controllerHwnd); err != nil {
		a.logger.Printf("tray add failed: %v", err)
	}
	return nil
}

func (a *App) refreshSnapshot() error {
	snapshot, err := a.inventory.Snapshot()
	if err != nil {
		return err
	}
	a.lastSnapshot = snapshot
	return nil
}

func (a *App) shutdown() {
	a.shuttingDown.Store(true)
	if a.hook != nil {
		_ = a.hook.Close()
	}
	if a.watcher != nil {
		_ = a.watcher.Close()
	}
	if a.tray != nil && a.controllerHwnd != 0 {
		_ = a.tray.Delete(a.controllerHwnd)
	}
	if a.overlay != nil {
		a.overlay.Hide()
	}
	a.thumbnailWarmWG.Wait()
	if a.thumbnails != nil {
		a.thumbnails.Close()
	}
	if a.icons != nil {
		a.icons.Close()
	}
	if a.desktop != nil {
		a.desktop.Close()
	}
	if a.overlayHwnd != 0 {
		win32.DestroyWindow(a.overlayHwnd)
		a.overlayHwnd = 0
	}
	if a.controllerHwnd != 0 {
		win32.DestroyWindow(a.controllerHwnd)
		a.controllerHwnd = 0
	}
}

func (a *App) loop() error {
	var msg win32.MSG
	for {
		ok, err := win32.GetMessage(&msg, 0, 0, 0)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessage(&msg)
	}
}

func (a *App) controllerWndProc(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case msgShutdownRequested:
		win32.PostQuitMessage(0)
		return 0
	case msgHookTabPressed:
		if a.shuttingDown.Load() {
			return 0
		}
		a.onTabPressed()
		return 0
	case msgHookAltReleased:
		a.onAltReleased()
		return 0
	case msgHookCancel:
		a.cancelSession()
		return 0
	case msgForegroundChanged:
		a.onForegroundChanged(wParam)
		return 0
	case msgThumbnailsReady:
		if !a.shuttingDown.Load() && a.session.State == session.StateCycling {
			a.overlay.RefreshThumbnails()
		}
		return 0
	case msgIconsReady:
		if !a.shuttingDown.Load() && a.session.State == session.StateCycling {
			a.overlay.Refresh()
		}
		return 0
	case win32.WM_COMMAND:
		if uint32(wParam&0xffff) == ui.CommandExit {
			win32.PostQuitMessage(0)
			return 0
		}
	case win32.WM_DESTROY:
		win32.PostQuitMessage(0)
		return 0
	default:
		if msg == a.taskbarMsg {
			_ = a.tray.Add(hwnd)
			return 0
		}
		if msg == msgTray {
			switch trayNotificationCode(lParam) {
			case wmContextMenu, win32.WM_RBUTTONUP, ninSelect, ninKeySelect:
				a.tray.ShowMenu(hwnd, ui.CommandExit)
			}
			return 0
		}
	}
	return win32.DefWindowProc(hwnd, msg, wParam, lParam)
}

func (a *App) overlayWndProc(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == win32.WM_PAINT {
		a.overlay.Paint(hwnd, a.icons, a.thumbnails)
		return 0
	}
	return win32.DefWindowProc(hwnd, msg, wParam, lParam)
}

func (a *App) onForegroundChanged(hwnd uintptr) {
	if a.session.State == session.StateCycling || a.session.State == session.StateCommitPending {
		return
	}
	id := windows.WindowID(hwnd)
	if id == 0 {
		return
	}
	if a.validSwitchTarget(id) {
		a.rememberSnapshotTarget(id)
		a.mru.MoveToFront(id)
	}
}

func (a *App) onTabPressed() {
	if a.session.State == session.StateCycling {
		a.session.Advance()
		if !a.previewSelection() {
			a.cancelSession()
			return
		}
		a.overlay.UpdateSelection(a.session.SelectedIndex)
		return
	}
	a.pruneCachedSnapshot()
	if len(a.lastSnapshot.Order) < 2 {
		if err := a.refreshSnapshot(); err != nil {
			a.logger.Printf("refresh snapshot: %v", err)
			return
		}
	}
	candidates := a.mru.BuildCandidates(a.lastSnapshot)
	startedFrom := windows.WindowID(win32.GetForegroundWindow())
	if !a.session.Start(candidates, startedFrom) {
		return
	}
	if !a.previewSelection() {
		a.session.Reset()
		return
	}
	items, metrics := a.renderOverlay()
	a.warmSessionIconsAsync(items)
	a.warmSessionThumbnailsAsync(items, metrics)
}

func (a *App) pruneCachedSnapshot() {
	if len(a.lastSnapshot.Order) == 0 || len(a.lastSnapshot.ByID) == 0 {
		a.lastSnapshot.Order = a.lastSnapshot.Order[:0]
		return
	}
	kept := a.lastSnapshot.Order[:0]
	for _, id := range a.lastSnapshot.Order {
		if _, ok := a.lastSnapshot.ByID[id]; !ok || !a.validSwitchTarget(id) {
			delete(a.lastSnapshot.ByID, id)
			continue
		}
		kept = append(kept, id)
	}
	a.lastSnapshot.Order = kept
}

func (a *App) rememberSnapshotTarget(id windows.WindowID) {
	if id == 0 {
		return
	}
	if a.lastSnapshot.ByID == nil {
		a.lastSnapshot.ByID = make(map[windows.WindowID]windows.WindowInfo)
	}
	if _, ok := a.lastSnapshot.ByID[id]; !ok {
		a.lastSnapshot.Order = append(a.lastSnapshot.Order, id)
	}
	info := a.lastSnapshot.ByID[id]
	if info.ID == 0 {
		info.ID = id
	}
	info.IsAppWindow = true
	a.lastSnapshot.ByID[id] = info
}

func (a *App) renderOverlay() ([]windows.WindowInfo, ui.OverlayMetrics) {
	items := a.sessionItems()
	anchor := a.session.StartedFrom.HWND()
	if anchor == 0 {
		anchor = win32.GetForegroundWindow()
	}
	metrics := ui.ComputeMetricsForAnchor(anchor, len(items))
	a.overlay.UpdateWithMetrics(anchor, items, a.session.SelectedIndex, metrics)
	return items, metrics
}

func (a *App) sessionItems() []windows.WindowInfo {
	items := make([]windows.WindowInfo, 0, len(a.session.Candidates))
	for _, id := range a.session.Candidates {
		info, ok := a.lastSnapshot.ByID[id]
		if !ok {
			continue
		}
		items = append(items, info)
	}
	return items
}

func (a *App) warmSessionThumbnailsAsync(items []windows.WindowInfo, metrics ui.OverlayMetrics) {
	if a.thumbnails == nil || a.shuttingDown.Load() {
		return
	}
	controller := a.controllerHwnd
	a.thumbnailWarmWG.Add(1)
	go func() {
		defer a.thumbnailWarmWG.Done()
		if a.shuttingDown.Load() {
			return
		}
		a.thumbnails.Warm(items, metrics.ThumbnailWidth, metrics.ThumbnailHeight)
		if !a.shuttingDown.Load() && controller != 0 {
			win32.PostMessage(controller, msgThumbnailsReady, 0, 0)
		}
	}()
}

func (a *App) warmSessionIconsAsync(items []windows.WindowInfo) {
	if a.icons == nil || a.shuttingDown.Load() {
		return
	}
	controller := a.controllerHwnd
	copied := append([]windows.WindowInfo(nil), items...)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := win32.CoInitialize(); err != nil {
			if a.logger != nil {
				a.logger.Printf("initialize COM for icon warm: %v", err)
			}
			return
		}
		defer win32.CoUninitialize()

		if a.shuttingDown.Load() {
			return
		}
		if a.icons.Warm(copied) && !a.shuttingDown.Load() && controller != 0 {
			win32.PostMessage(controller, msgIconsReady, 0, 0)
		}
	}()
}

func (a *App) onAltReleased() {
	selected, ok := a.session.BeginCommit()
	if !ok {
		return
	}
	a.overlay.Hide()
	if err := a.finalizeSelection(selected); err != nil {
		a.logger.Printf("commit failed: %v", err)
	}
	if err := a.releaseAltModifier(); err != nil {
		a.logger.Printf("release Alt modifier failed: %v", err)
	}
	a.session.Reset()
}

func (a *App) previewSelection() bool {
	index, selected, err := a.activateCandidateFrom(a.session.SelectedIndex, 0)
	if err != nil {
		a.logger.Printf("preview activation failed: %v", err)
		return false
	}
	a.session.SelectedIndex = index
	a.rememberSnapshotTarget(selected)
	return true
}

func (a *App) finalizeSelection(selected windows.WindowID) error {
	if selected != 0 && a.validSwitchTarget(selected) {
		a.mru.MoveToFront(selected)
		return nil
	}

	index, candidate, err := a.activateCandidateFrom(a.session.SelectedIndex+1, selected)
	if err != nil {
		return fmt.Errorf("finalize selected target %v: %w", selected, err)
	}
	a.session.SelectedIndex = index
	a.mru.MoveToFront(candidate)
	return nil
}

func (a *App) activateCandidateFrom(start int, skip windows.WindowID) (int, windows.WindowID, error) {
	if len(a.session.Candidates) == 0 {
		return 0, 0, fmt.Errorf("no session candidates")
	}

	start = normalizeCandidateIndex(start, len(a.session.Candidates))
	var firstErr error
	for i := 0; i < len(a.session.Candidates); i++ {
		index := (start + i) % len(a.session.Candidates)
		candidate := a.session.Candidates[index]
		if candidate == 0 || candidate == skip || !a.validSwitchTarget(candidate) {
			continue
		}
		if err := a.activate(candidate); err == nil {
			return index, candidate, nil
		} else if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return 0, 0, firstErr
	}
	return 0, 0, fmt.Errorf("no valid switch target")
}

func (a *App) cancelSession() {
	startedFrom := a.session.StartedFrom
	a.session.Cancel()
	a.overlay.Hide()
	if startedFrom != 0 && a.validSwitchTarget(startedFrom) {
		if err := a.activate(startedFrom); err != nil {
			a.logger.Printf("restore starting window failed: %v", err)
		}
	}
	a.session.Reset()
}

func (a *App) activate(target windows.WindowID) error {
	if a.activateTarget != nil {
		return a.activateTarget(target)
	}
	return windows.Activate(target)
}

func (a *App) releaseAltModifier() error {
	if a.releaseModifiers != nil {
		return a.releaseModifiers()
	}
	return win32.SendForegroundUnlockInput()
}

func (a *App) validSwitchTarget(target windows.WindowID) bool {
	if a.isValidSwitchTarget != nil {
		return a.isValidSwitchTarget(target)
	}
	if a.inventory == nil {
		return false
	}
	return a.inventory.IsValidSwitchTarget(target)
}

func normalizeCandidateIndex(start, count int) int {
	if count <= 0 {
		return 0
	}
	start %= count
	if start < 0 {
		start += count
	}
	return start
}

func trayNotificationCode(lParam uintptr) uint32 {
	return uint32(lParam & 0xffff)
}

func (a *App) requestShutdown() {
	if a.controllerHwnd == 0 {
		return
	}
	win32.PostMessage(a.controllerHwnd, msgShutdownRequested, 0, 0)
}
