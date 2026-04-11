package windows

import (
	"errors"
	"fmt"

	"better_alt_tab/internal/win32"
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
		if !win32.ShowWindowAsync(hwnd, win32.SW_RESTORE) {
			win32.ShowWindow(hwnd, win32.SW_RESTORE)
		}
	}
	if win32.SetForegroundWindow(hwnd) {
		return nil
	}
	if err := win32.SendForegroundUnlockInput(); err != nil {
		return fmt.Errorf("send unlock input after direct foreground failed: %w", err)
	}
	if !win32.SetForegroundWindow(hwnd) {
		return fmt.Errorf("set foreground failed after unlock input")
	}
	return nil
}
