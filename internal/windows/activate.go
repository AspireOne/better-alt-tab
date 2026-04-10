package windows

import (
	"errors"
	"fmt"

	"quick_app_switcher/internal/win32"
)

func Activate(target WindowID) error {
	if target == 0 {
		return errors.New("empty target")
	}
	hwnd := target.HWND()
	if !win32.IsWindow(hwnd) {
		return fmt.Errorf("window %v no longer exists", hwnd)
	}
	if win32.IsIconic(hwnd) {
		win32.ShowWindow(hwnd, win32.SW_RESTORE)
	}
	if err := win32.SendForegroundUnlockInput(); err != nil {
		return fmt.Errorf("send unlock input: %w", err)
	}
	if !win32.SetForegroundWindow(hwnd) {
		return fmt.Errorf("set foreground failed")
	}
	if current := win32.GetForegroundWindow(); current != hwnd {
		return fmt.Errorf("foreground verification failed: got %v want %v", current, hwnd)
	}
	return nil
}
