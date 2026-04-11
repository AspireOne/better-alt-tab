package ui

import (
	"fmt"

	"better_alt_tab/internal/config"
	"better_alt_tab/internal/win32"
)

const (
	settingsWidth  int32 = 320
	settingsHeight int32 = 188

	controlShowThumbnails  uint32 = 2001
	controlLaunchOnStartup uint32 = 2002
	controlInstantPreview  uint32 = 2003
	controlSave            uint32 = 2004
	controlCancel          uint32 = 2005
)

type SettingsWindow struct {
	hwnd            win32.HWND
	showThumbnails  win32.HWND
	launchOnStartup win32.HWND
	instantPreview  win32.HWND
}

func NewSettingsWindow() *SettingsWindow {
	return &SettingsWindow{}
}

func (w *SettingsWindow) Hwnd() win32.HWND {
	return w.hwnd
}

func (w *SettingsWindow) Create(instance win32.HINSTANCE, className string, anchor win32.HWND) error {
	if w.hwnd != 0 && win32.IsWindow(w.hwnd) {
		return nil
	}

	x, y := settingsWindowPosition(anchor)
	hwnd, err := win32.CreateWindowAt(
		win32.WS_EX_TOOLWINDOW,
		win32.WS_OVERLAPPED|win32.WS_CAPTION|win32.WS_SYSMENU,
		className,
		"Quick App Switcher Settings",
		x,
		y,
		settingsWidth,
		settingsHeight,
		0,
		0,
		instance,
		0,
	)
	if err != nil {
		return err
	}

	w.hwnd = hwnd
	if err := w.createControls(instance); err != nil {
		w.Destroy()
		return err
	}
	return nil
}

func (w *SettingsWindow) Show(cfg config.Config) {
	if w.hwnd == 0 {
		return
	}
	w.SetConfig(cfg)
	win32.ShowWindow(w.hwnd, win32.SW_RESTORE)
	win32.SetForegroundWindow(w.hwnd)
}

func (w *SettingsWindow) Hide() {
	if w.hwnd == 0 {
		return
	}
	win32.ShowWindow(w.hwnd, win32.SW_HIDE)
}

func (w *SettingsWindow) Destroy() {
	if w.hwnd == 0 {
		return
	}
	win32.DestroyWindow(w.hwnd)
	w.hwnd = 0
	w.showThumbnails = 0
	w.launchOnStartup = 0
	w.instantPreview = 0
}

func (w *SettingsWindow) SetConfig(cfg config.Config) {
	win32.SetCheckboxChecked(w.showThumbnails, cfg.ShowThumbnails)
	win32.SetCheckboxChecked(w.launchOnStartup, cfg.LaunchOnStartup)
	win32.SetCheckboxChecked(w.instantPreview, cfg.InstantSwitchPreview)
}

func (w *SettingsWindow) Config() config.Config {
	return config.Config{
		ShowThumbnails:       win32.CheckboxChecked(w.showThumbnails),
		LaunchOnStartup:      win32.CheckboxChecked(w.launchOnStartup),
		InstantSwitchPreview: win32.CheckboxChecked(w.instantPreview),
	}
}

func (w *SettingsWindow) HandleCommand(command uint32, onSave func() error, onCancel func()) (bool, error) {
	switch command {
	case controlSave:
		if onSave != nil {
			return true, onSave()
		}
		return true, nil
	case controlCancel:
		if onCancel != nil {
			onCancel()
		}
		return true, nil
	default:
		return false, nil
	}
}

func (w *SettingsWindow) createControls(instance win32.HINSTANCE) error {
	var err error
	if w.showThumbnails, err = createCheckbox(w.hwnd, instance, 16, 16, 260, 24, controlShowThumbnails, "Show window thumbnails"); err != nil {
		return fmt.Errorf("create show thumbnails checkbox: %w", err)
	}
	if w.launchOnStartup, err = createCheckbox(w.hwnd, instance, 16, 48, 260, 24, controlLaunchOnStartup, "Launch on Windows startup"); err != nil {
		return fmt.Errorf("create launch on startup checkbox: %w", err)
	}
	if w.instantPreview, err = createCheckbox(w.hwnd, instance, 16, 80, 280, 24, controlInstantPreview, "Preview selected window while cycling"); err != nil {
		return fmt.Errorf("create instant preview checkbox: %w", err)
	}
	if _, err = createButton(w.hwnd, instance, 132, 122, 76, 28, controlSave, "Save", true); err != nil {
		return fmt.Errorf("create save button: %w", err)
	}
	if _, err = createButton(w.hwnd, instance, 220, 122, 76, 28, controlCancel, "Cancel", false); err != nil {
		return fmt.Errorf("create cancel button: %w", err)
	}
	return nil
}

func createCheckbox(parent win32.HWND, instance win32.HINSTANCE, x, y, width, height int32, id uint32, text string) (win32.HWND, error) {
	return win32.CreateWindowAt(
		0,
		win32.WS_CHILD|win32.WS_VISIBLE|win32.WS_TABSTOP|win32.BS_AUTOCHECKBOX,
		"Button",
		text,
		x,
		y,
		width,
		height,
		parent,
		win32.HMENU(id),
		instance,
		0,
	)
}

func createButton(parent win32.HWND, instance win32.HINSTANCE, x, y, width, height int32, id uint32, text string, isDefault bool) (win32.HWND, error) {
	style := uint32(win32.WS_CHILD | win32.WS_VISIBLE | win32.WS_TABSTOP | win32.BS_PUSHBUTTON)
	if isDefault {
		style = win32.WS_CHILD | win32.WS_VISIBLE | win32.WS_TABSTOP | win32.BS_DEFPUSHBUTTON
	}
	return win32.CreateWindowAt(
		0,
		style,
		"Button",
		text,
		x,
		y,
		width,
		height,
		parent,
		win32.HMENU(id),
		instance,
		0,
	)
}

func settingsWindowPosition(anchor win32.HWND) (int32, int32) {
	if anchor == 0 {
		anchor = win32.GetForegroundWindow()
	}
	monitor := win32.MonitorFromWindow(anchor)
	workArea := win32.GetMonitorRect(monitor)
	return workArea.Left + (workArea.Right-workArea.Left-settingsWidth)/2,
		workArea.Top + (workArea.Bottom-workArea.Top-settingsHeight)/2
}
