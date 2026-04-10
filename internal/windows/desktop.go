package windows

import "quick_app_switcher/internal/win32"

type DesktopManager struct {
	manager   *win32.VirtualDesktopManager
	available bool
}

func NewDesktopManager() (*DesktopManager, error) {
	manager, err := win32.NewVirtualDesktopManager()
	if err != nil {
		return &DesktopManager{}, err
	}
	return &DesktopManager{manager: manager, available: true}, nil
}

func (d *DesktopManager) Close() {
	if d != nil && d.manager != nil {
		d.manager.Release()
	}
}

func (d *DesktopManager) IsWindowOnCurrentDesktop(hwnd win32.HWND) bool {
	if d == nil || !d.available || d.manager == nil {
		return false
	}
	ok, err := d.manager.IsWindowOnCurrentDesktop(hwnd)
	if err != nil {
		return false
	}
	return ok
}
