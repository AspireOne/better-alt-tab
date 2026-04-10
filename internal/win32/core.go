package win32

import (
	"errors"
	"syscall"
	"unsafe"
)

func utf16Ptr(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

func ignoreSyscall3(_, _ uintptr, _ error) {}

func lastErr() error {
	r, _, _ := procGetLastError.Call()
	if r == 0 {
		return nil
	}
	return syscall.Errno(r)
}

func CreateNamedMutex(name string) (Handle, bool, error) {
	r, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(utf16Ptr(name))))
	if r == 0 {
		return 0, false, err
	}
	return Handle(r), errors.Is(lastErr(), syscall.Errno(183)), nil
}

func CloseHandle(handle Handle) error {
	r, _, err := procCloseHandle.Call(uintptr(handle))
	if r == 0 {
		return err
	}
	return nil
}

func GetCurrentThreadID() uint32 {
	r, _, _ := procGetCurrentThreadId.Call()
	return uint32(r)
}

func RegisterWindowClass(className string, wndProc uintptr, instance HINSTANCE, icon HICON) (uint16, error) {
	cursor, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
	wc := WNDCLASSEX{
		Size:      uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:     CS_HREDRAW | CS_VREDRAW,
		WndProc:   wndProc,
		Instance:  instance,
		Icon:      icon,
		Cursor:    HCURSOR(cursor),
		ClassName: utf16Ptr(className),
	}
	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return 0, err
	}
	return uint16(r), nil
}

func CreateWindow(exStyle, style uint32, className, windowName string, instance HINSTANCE, lpParam uintptr) (HWND, error) {
	r, _, err := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(windowName))),
		uintptr(style),
		0, 0, 0, 0,
		0, 0,
		uintptr(instance),
		lpParam,
	)
	if r == 0 {
		return 0, err
	}
	return HWND(r), nil
}

func DefWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func DestroyWindow(hwnd HWND) {
	ignoreSyscall3(procDestroyWindow.Call(uintptr(hwnd)))
}

func GetMessage(msg *MSG, hwnd HWND, min, max uint32) (bool, error) {
	r, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(msg)), uintptr(hwnd), uintptr(min), uintptr(max))
	switch int32(r) {
	case -1:
		if err != syscall.Errno(0) {
			return false, err
		}
		return false, lastErr()
	case 0:
		return false, nil
	default:
		return true, nil
	}
}

func TranslateMessage(msg *MSG) {
	ignoreSyscall3(procTranslateMessage.Call(uintptr(unsafe.Pointer(msg))))
}

func DispatchMessage(msg *MSG) {
	ignoreSyscall3(procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg))))
}

func PostQuitMessage(code int32) {
	ignoreSyscall3(procPostQuitMessage.Call(uintptr(code)))
}

func PostMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) bool {
	r, _, _ := procPostMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r != 0
}

func PostThreadMessage(threadID uint32, msg uint32, wParam, lParam uintptr) bool {
	r, _, _ := procPostThreadMessageW.Call(uintptr(threadID), uintptr(msg), wParam, lParam)
	return r != 0
}

func SetKeyboardHook(callback uintptr) (Handle, error) {
	r, _, err := procSetWindowsHookExW.Call(WH_KEYBOARD_LL, callback, 0, 0)
	if r == 0 {
		return 0, err
	}
	return Handle(r), nil
}

func CallNextHook(hook Handle, code int32, wParam, lParam uintptr) uintptr {
	r, _, _ := procCallNextHookEx.Call(uintptr(hook), uintptr(code), wParam, lParam)
	return r
}

func UnhookWindowsHook(hook Handle) error {
	r, _, err := procUnhookWindowsHookEx.Call(uintptr(hook))
	if r == 0 {
		return err
	}
	return nil
}

func EnumWindows(cb func(HWND) bool) error {
	callback := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		if cb(HWND(hwnd)) {
			return 1
		}
		return 0
	})
	r, _, err := procEnumWindows.Call(callback, 0)
	if r == 0 && err != syscall.Errno(0) {
		return err
	}
	return nil
}

func IsWindow(hwnd HWND) bool {
	r, _, _ := procIsWindow.Call(uintptr(hwnd))
	return r != 0
}

func IsWindowVisible(hwnd HWND) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

func IsIconic(hwnd HWND) bool {
	r, _, _ := procIsIconic.Call(uintptr(hwnd))
	return r != 0
}

func GetWindowText(hwnd HWND) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	ignoreSyscall3(procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))))
	return syscall.UTF16ToString(buf)
}

func GetWindowProcessID(hwnd HWND) uint32 {
	var pid uint32
	ignoreSyscall3(procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid))))
	return pid
}

func GetWindowStyle(hwnd HWND) uintptr {
	index := int32(GWL_STYLE)
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index))
	return r
}

func GetWindowExStyle(hwnd HWND) uintptr {
	index := int32(GWL_EXSTYLE)
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index))
	return r
}

func GetWindow(hwnd HWND, cmd uint32) HWND {
	r, _, _ := procGetWindow.Call(uintptr(hwnd), uintptr(cmd))
	return HWND(r)
}

func GetAncestor(hwnd HWND, flags uint32) HWND {
	r, _, _ := procGetAncestor.Call(uintptr(hwnd), uintptr(flags))
	return HWND(r)
}

func GetLastActivePopup(hwnd HWND) HWND {
	r, _, _ := procGetLastActivePopup.Call(uintptr(hwnd))
	return HWND(r)
}

func ShowWindow(hwnd HWND, cmd int32) bool {
	r, _, _ := procShowWindow.Call(uintptr(hwnd), uintptr(cmd))
	return r != 0
}

func SetForegroundWindow(hwnd HWND) bool {
	r, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
	return r != 0
}

func GetForegroundWindow() HWND {
	r, _, _ := procGetForegroundWindow.Call()
	return HWND(r)
}
