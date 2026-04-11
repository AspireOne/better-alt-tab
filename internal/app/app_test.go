package app

import (
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"better_alt_tab/internal/config"
	"better_alt_tab/internal/mru"
	"better_alt_tab/internal/session"
	"better_alt_tab/internal/theme"
	"better_alt_tab/internal/ui"
	"better_alt_tab/internal/win32"
	"better_alt_tab/internal/windows"
)

func TestTrayNotificationCodeUsesLowWord(t *testing.T) {
	got := trayNotificationCode(0x1234007B)
	if got != wmContextMenu {
		t.Fatalf("got %#x want %#x", got, wmContextMenu)
	}
}

func TestHandleCommandOpensSettings(t *testing.T) {
	a := newTestApp()
	calls := 0
	a.openSettings = func() error {
		calls++
		return nil
	}

	handled := a.handleCommand(ui.CommandOpenSettings)
	if !handled {
		t.Fatal("expected open settings command to be handled")
	}
	if calls != 1 {
		t.Fatalf("got open settings calls %d want 1", calls)
	}
}

func TestHandleCommandOpensConfigFile(t *testing.T) {
	a := newTestApp()
	calls := 0
	a.openConfigFile = func() error {
		calls++
		return nil
	}

	handled := a.handleCommand(ui.CommandOpenConfigFile)
	if !handled {
		t.Fatal("expected open config file command to be handled")
	}
	if calls != 1 {
		t.Fatalf("got open config file calls %d want 1", calls)
	}
}

func TestHandleCommandReloadsTheme(t *testing.T) {
	a := newTestApp()
	calls := 0
	a.loadTheme = func(string) (theme.Theme, error) {
		calls++
		return theme.Default(), nil
	}

	handled := a.handleCommand(ui.CommandReloadTheme)
	if !handled {
		t.Fatal("expected reload theme command to be handled")
	}
	if calls != 1 {
		t.Fatalf("got reload theme calls %d want 1", calls)
	}
}

func TestHandleCommandIgnoresUnknownCommand(t *testing.T) {
	a := newTestApp()

	if a.handleCommand(9999) {
		t.Fatal("expected unknown command to be ignored")
	}
}

func TestNormalizeCandidateIndex(t *testing.T) {
	tests := []struct {
		name  string
		start int
		count int
		want  int
	}{
		{name: "empty", start: 4, count: 0, want: 0},
		{name: "in range", start: 1, count: 3, want: 1},
		{name: "wraps forward", start: 4, count: 3, want: 1},
		{name: "wraps negative", start: -1, count: 3, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCandidateIndex(tt.start, tt.count); got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestActivateCandidateFromSkipsInvalidAndFailedTargets(t *testing.T) {
	activated := make([]windows.WindowID, 0, 2)
	a := newTestApp()
	a.session.Candidates = []windows.WindowID{10, 20, 30}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		if target == 10 {
			return errors.New("activate failed")
		}
		return nil
	}

	index, target, err := a.activateCandidateFrom(0, 0)
	if err != nil {
		t.Fatalf("activateCandidateFrom returned error: %v", err)
	}
	if index != 2 {
		t.Fatalf("got index %d want 2", index)
	}
	if target != 30 {
		t.Fatalf("got target %v want 30", target)
	}
	if !reflect.DeepEqual(activated, []windows.WindowID{10, 30}) {
		t.Fatalf("got activations %v want [10 30]", activated)
	}
}

func TestPreviewSelectionUpdatesIndexAndSnapshotTarget(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.lastSnapshot = windows.InventorySnapshot{
		ByID: map[windows.WindowID]windows.WindowInfo{
			10: {ID: 10},
			20: {ID: 20},
			30: {ID: 30},
		},
	}
	var activated windows.WindowID
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }
	a.activateTarget = func(target windows.WindowID) error {
		activated = target
		return nil
	}

	if !a.previewSelection() {
		t.Fatal("expected previewSelection to succeed")
	}
	if activated != 20 {
		t.Fatalf("got activated %v want 20", activated)
	}
	if a.session.SelectedIndex != 1 {
		t.Fatalf("got selected index %d want 1", a.session.SelectedIndex)
	}
	if _, ok := a.lastSnapshot.ByID[20]; !ok {
		t.Fatal("expected selected target to be remembered in snapshot")
	}
}

func TestOnTabPressedWithoutInstantPreviewOnlyUpdatesSelection(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onTabPressed()

	if !reflect.DeepEqual(activated, []windows.WindowID(nil)) {
		t.Fatalf("got activations %v want none", activated)
	}
	if a.session.SelectedIndex != 2 {
		t.Fatalf("got selected index %d want 2", a.session.SelectedIndex)
	}
}

func TestFinalizeSelectionFallsBackToNextActivatableTarget(t *testing.T) {
	a := newTestApp()
	a.mru = mru.New()
	a.session = session.SwitchSession{
		State:         session.StateCommitPending,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}
	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	if err := a.finalizeSelection(20); err != nil {
		t.Fatalf("finalizeSelection returned error: %v", err)
	}
	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
	got := a.mru.BuildCandidates(windows.InventorySnapshot{
		Order: []windows.WindowID{10, 20, 30},
	})
	want := []windows.WindowID{30, 10, 20}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got MRU order %v want %v", got, want)
	}
}

func TestCancelSessionRestoresStartingWindowWhenValid(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:          session.StateCycling,
		Candidates:     []windows.WindowID{10, 20},
		SelectedIndex:  1,
		StartedFrom:    10,
		OverlayVisible: true,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return target == 10 }
	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.cancelSession()

	if !reflect.DeepEqual(activated, []windows.WindowID{10}) {
		t.Fatalf("got activations %v want [10]", activated)
	}
	if a.session.State != session.StateIdle {
		t.Fatalf("got state %v want idle", a.session.State)
	}
	if a.session.OverlayVisible {
		t.Fatal("expected overlay hidden after cancel")
	}
}

func TestOnAltReleasedFinalizesSelectionAndReleasesModifiers(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	releaseCalls := 0
	a.releaseModifiers = func() error {
		releaseCalls++
		return nil
	}

	a.onAltReleased()

	if releaseCalls != 1 {
		t.Fatalf("got release calls %d want 1", releaseCalls)
	}
	if a.session.State != session.StateIdle {
		t.Fatalf("got state %v want idle", a.session.State)
	}
}

func TestOnAltReleasedWithoutInstantPreviewActivatesSelectedTarget(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 2,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onAltReleased()

	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
}

func TestOnAltReleasedWithoutInstantPreviewFallsBackToNextValidTarget(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onAltReleased()

	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
}

func TestCancelSessionWithoutInstantPreviewDoesNotRestoreStartingWindow(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:          session.StateCycling,
		Candidates:     []windows.WindowID{10, 20},
		SelectedIndex:  1,
		StartedFrom:    10,
		OverlayVisible: true,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.cancelSession()

	if len(activated) != 0 {
		t.Fatalf("got activations %v want none", activated)
	}
}

func TestSaveSettingsConfigUpdatesRuntimeAndStartup(t *testing.T) {
	a := newTestApp()
	want := config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: false,
		Theme:                "sunset",
	}
	saved := config.Config{}
	startupCalls := 0
	a.saveConfig = func(cfg config.Config) error {
		saved = cfg
		return nil
	}
	a.syncStartup = func(enabled bool) error {
		startupCalls++
		if !enabled {
			t.Fatal("expected startup sync to receive enabled=true")
		}
		return nil
	}

	if err := a.saveSettingsConfig(want); err != nil {
		t.Fatalf("saveSettingsConfig returned error: %v", err)
	}
	if saved != want {
		t.Fatalf("got saved config %+v want %+v", saved, want)
	}
	if a.cfg != want {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, want)
	}
	if startupCalls != 1 {
		t.Fatalf("got startup calls %d want 1", startupCalls)
	}
}

func TestSaveSettingsConfigDoesNotPersistWhenThemeLoadFails(t *testing.T) {
	a := newTestApp()
	original := a.cfg
	a.loadTheme = func(string) (theme.Theme, error) {
		return theme.Theme{}, errors.New("bad theme")
	}

	calls := 0
	a.saveConfig = func(cfg config.Config) error {
		calls++
		return nil
	}

	err := a.saveSettingsConfig(config.Config{Theme: "broken"})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 0 {
		t.Fatalf("got save calls %d want 0", calls)
	}
	if a.cfg != original {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, original)
	}
	if a.theme != theme.Default() {
		t.Fatalf("got runtime theme %+v want default theme", a.theme)
	}
}

func TestSaveSettingsConfigRollsBackThemeWhenSaveFails(t *testing.T) {
	a := newTestApp()
	oldTheme := theme.Default()
	oldTheme.Window.Opacity = 200
	newTheme := theme.Default()
	newTheme.Window.Opacity = 180
	a.theme = oldTheme

	a.loadTheme = func(name string) (theme.Theme, error) {
		if name != "sunset" {
			t.Fatalf("got theme name %q want %q", name, "sunset")
		}
		return newTheme, nil
	}

	calls := 0
	a.saveConfig = func(cfg config.Config) error {
		calls++
		return errors.New("disk full")
	}

	err := a.saveSettingsConfig(config.Config{Theme: "sunset"})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("got save calls %d want 1", calls)
	}
	if a.theme != oldTheme {
		t.Fatalf("got runtime theme %+v want %+v", a.theme, oldTheme)
	}
	if a.cfg != config.Default() {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, config.Default())
	}
}

func TestSaveSettingsPreservesTheme(t *testing.T) {
	const themeName = "sunset"

	a := newTestApp()
	a.cfg.Theme = themeName
	a.settings = ui.NewSettingsWindow()
	a.settings.SetConfig(config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: false,
	})

	var saved config.Config
	a.saveConfig = func(cfg config.Config) error {
		saved = cfg
		return nil
	}
	a.loadTheme = func(name string) (theme.Theme, error) {
		if name != themeName {
			t.Fatalf("got theme name %q want %q", name, themeName)
		}
		return theme.Default(), nil
	}
	a.syncStartup = func(bool) error { return nil }

	if err := a.saveSettings(); err != nil {
		t.Fatalf("saveSettings returned error: %v", err)
	}
	if saved.Theme != themeName {
		t.Fatalf("got saved theme %q want %q", saved.Theme, themeName)
	}
}

func TestLoadCurrentConfigUpdatesRuntimeWithoutSyncingStartup(t *testing.T) {
	a := newTestApp()
	want := config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: true,
		Theme:                "sunset",
	}
	a.loadConfig = func() (config.Config, error) {
		return want, nil
	}
	a.loadTheme = func(name string) (theme.Theme, error) {
		if name != "sunset" {
			t.Fatalf("got theme name %q want %q", name, "sunset")
		}
		return theme.Default(), nil
	}
	startupCalls := 0
	a.syncStartup = func(bool) error {
		startupCalls++
		return nil
	}

	got, err := a.loadCurrentConfig()
	if err != nil {
		t.Fatalf("loadCurrentConfig returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got loaded config %+v want %+v", got, want)
	}
	if a.cfg != want {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, want)
	}
	if startupCalls != 0 {
		t.Fatalf("got startup calls %d want 0", startupCalls)
	}
}

func TestLoadSettingsConfigContinuesWhenThemeLoadFails(t *testing.T) {
	a := newTestApp()
	want := config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: true,
		Theme:                "broken",
	}
	a.loadConfig = func() (config.Config, error) {
		return want, nil
	}
	a.loadTheme = func(name string) (theme.Theme, error) {
		if name != "broken" {
			t.Fatalf("got theme name %q want %q", name, "broken")
		}
		return theme.Default(), errors.New("theme parse failed")
	}

	got, themeErr, err := a.loadSettingsConfig()
	if err != nil {
		t.Fatalf("loadSettingsConfig returned config error: %v", err)
	}
	if themeErr == nil {
		t.Fatal("expected theme error")
	}
	if got != want {
		t.Fatalf("got loaded config %+v want %+v", got, want)
	}
	if a.cfg != want {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, want)
	}
}

func TestHandleInterruptsRequestsShutdownOnFirstInterrupt(t *testing.T) {
	a := newTestApp()
	requested := make(chan struct{}, 1)
	a.controllerHwnd = 1
	a.forceExit = func(code int) {
		t.Fatalf("unexpected forced exit with code %d", code)
	}
	originalPost := win32PostMessage
	win32PostMessage = func(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) bool {
		if hwnd != a.controllerHwnd {
			t.Fatalf("got hwnd %v want %v", hwnd, a.controllerHwnd)
		}
		if msg != msgShutdownRequested {
			t.Fatalf("got message %#x want %#x", msg, msgShutdownRequested)
		}
		requested <- struct{}{}
		return true
	}
	defer func() {
		win32PostMessage = originalPost
	}()

	signals := make(chan os.Signal, 2)
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		a.handleInterrupts(signals, done)
	}()

	signals <- os.Interrupt

	select {
	case <-requested:
	case <-finished:
		t.Fatal("handler returned before processing first interrupt")
	}

	close(done)
	<-finished
}

func TestHandleInterruptsForcesExitOnSecondInterrupt(t *testing.T) {
	a := newTestApp()
	a.controllerHwnd = 1
	exitCode := make(chan int, 1)
	a.forceExit = func(code int) {
		exitCode <- code
	}
	originalPost := win32PostMessage
	win32PostMessage = func(hwnd win32.HWND, msg uint32, wParam, lParam uintptr) bool {
		return true
	}
	defer func() {
		win32PostMessage = originalPost
	}()

	signals := make(chan os.Signal, 2)
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		a.handleInterrupts(signals, done)
	}()

	signals <- os.Interrupt
	signals <- os.Interrupt

	select {
	case code := <-exitCode:
		if code != 130 {
			t.Fatalf("got exit code %d want 130", code)
		}
	case <-finished:
		t.Fatal("handler returned before forcing exit")
	}
}

func newTestApp() *App {
	a := &App{
		logger:  log.New(io.Discard, "", 0),
		cfg:     config.Default(),
		theme:   theme.Default(),
		mru:     mru.New(),
		overlay: ui.NewOverlay(0, theme.Default()),
	}
	a.openSettings = func() error { return nil }
	a.openConfigFile = func() error { return nil }
	a.loadTheme = func(string) (theme.Theme, error) { return theme.Default(), nil }
	return a
}
