package win32

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

func PositionWindowNoActivate(hwnd HWND, x, y, width, height int32, show bool) {
	flags := uintptr(SWP_NOACTIVATE)
	if show {
		flags |= SWP_SHOWWINDOW
	} else {
		flags |= SWP_HIDEWINDOW
	}
	ignoreSyscall3(procSetWindowPos.Call(uintptr(hwnd), 0, uintptr(x), uintptr(y), uintptr(width), uintptr(height), flags))
}

func SetLayeredWindowAlpha(hwnd HWND, alpha byte) error {
	r, _, err := procSetLayeredWindowAttributes.Call(uintptr(hwnd), 0, uintptr(alpha), LWA_ALPHA)
	if r == 0 {
		return err
	}
	return nil
}

func InvalidateRect(hwnd HWND) {
	ignoreSyscall3(procInvalidateRect.Call(uintptr(hwnd), 0, 1))
}

func LoadDefaultApplicationIcon() HICON {
	r, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
	return HICON(r)
}

func SetForegroundEventHook(callback uintptr) (Handle, error) {
	r, _, err := procSetWinEventHook.Call(
		EVENT_SYSTEM_FOREGROUND,
		EVENT_SYSTEM_FOREGROUND,
		0,
		callback,
		0,
		0,
		WINEVENT_OUTOFCONTEXT,
	)
	if r == 0 {
		return 0, err
	}
	return Handle(r), nil
}

func UnhookWinEvent(hook Handle) error {
	r, _, err := procUnhookWinEvent.Call(uintptr(hook))
	if r == 0 {
		return err
	}
	return nil
}

func RegisterWindowMessage(name string) uint32 {
	r, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(utf16Ptr(name))))
	return uint32(r)
}

func CreateSolidBrush(color uintptr) HBRUSH {
	r, _, _ := procCreateSolidBrush.Call(color)
	return HBRUSH(r)
}

func DeleteObject(obj uintptr) {
	ignoreSyscall3(procDeleteObject.Call(obj))
}

func BeginPaint(hwnd HWND) (PAINTSTRUCT, HDC) {
	var ps PAINTSTRUCT
	r, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	return ps, HDC(r)
}

func EndPaint(hwnd HWND, ps *PAINTSTRUCT) {
	ignoreSyscall3(procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps))))
}

func FillRect(hdc HDC, rect *RECT, brush HBRUSH) {
	ignoreSyscall3(procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush)))
}

func DrawIconInRect(hdc HDC, rect RECT, icon HICON) {
	ignoreSyscall3(procDrawIconEx.Call(
		uintptr(hdc),
		uintptr(rect.Left),
		uintptr(rect.Top),
		uintptr(icon),
		uintptr(rect.Right-rect.Left),
		uintptr(rect.Bottom-rect.Top),
		0,
		0,
		3,
	))
}

func MonitorFromWindow(hwnd HWND) HMONITOR {
	r, _, _ := procMonitorFromWindow.Call(uintptr(hwnd), MONITOR_DEFAULTTONEAREST)
	return HMONITOR(r)
}

func GetMonitorRect(monitor HMONITOR) RECT {
	info := MONITORINFO{CbSize: uint32(unsafe.Sizeof(MONITORINFO{}))}
	ignoreSyscall3(procGetMonitorInfoW.Call(uintptr(monitor), uintptr(unsafe.Pointer(&info))))
	return info.RcWork
}

func SendForegroundUnlockInput() error {
	inputs := []INPUT{
		NewKeyboardInput(VK_MENU, 0, 0),
		NewKeyboardInput(VK_MENU, 0, KEYEVENTF_KEYUP),
	}
	r, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if r == 0 {
		return err
	}
	return nil
}

func GetClassName(hwnd HWND) string {
	buf := make([]uint16, 256)
	r, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:r])
}

func IsWindowCloaked(hwnd HWND) bool {
	var cloaked uint32
	r, _, _ := procDwmGetWindowAttribute.Call(uintptr(hwnd), DWMWA_CLOAKED, uintptr(unsafe.Pointer(&cloaked)), unsafe.Sizeof(cloaked))
	return r == 0 && cloaked != 0
}

const processQueryLimitedInformation = 0x1000

func GetWindowProcessPath(hwnd HWND) string {
	pid := GetWindowProcessID(hwnd)
	if pid == 0 {
		return ""
	}
	r, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if r == 0 {
		return ""
	}
	defer func() {
		_ = CloseHandle(Handle(r))
	}()
	buf := make([]uint16, MAX_PATH)
	n, _, _ := procGetModuleFileNameExW.Call(r, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return filepath.Clean(syscall.UTF16ToString(buf[:n]))
}

func AddTrayIcon(hwnd HWND, messageID uint32, icon HICON, tooltip string) error {
	var data NOTIFYICONDATA
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.HWnd = hwnd
	data.UID = 1
	data.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	data.UCallbackMessage = messageID
	data.HIcon = icon
	tooltipUTF16, _ := syscall.UTF16FromString(tooltip)
	copy(data.SzTip[:], tooltipUTF16)
	r, _, err := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&data)))
	if r == 0 {
		return err
	}
	data.UTimeoutOrVersion = NOTIFYICON_VERSION_4
	ignoreSyscall3(procShellNotifyIconW.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&data))))
	return nil
}

func DeleteTrayIcon(hwnd HWND) error {
	var data NOTIFYICONDATA
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.HWnd = hwnd
	data.UID = 1
	r, _, err := procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&data)))
	if r == 0 {
		return err
	}
	return nil
}

func ShowTrayMenu(hwnd HWND, commandID uint32) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer func() {
		ignoreSyscall3(procDestroyMenu.Call(menu))
	}()
	ignoreSyscall3(procAppendMenuW.Call(menu, MF_STRING, uintptr(commandID), uintptr(unsafe.Pointer(utf16Ptr("Close")))))
	var pt POINT
	ignoreSyscall3(procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))))
	cmd, _, _ := procTrackPopupMenu.Call(menu, TPM_RETURNCMD|TPM_NONOTIFY, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(hwnd), 0)
	if cmd != 0 {
		PostMessage(hwnd, WM_COMMAND, cmd, 0)
	}
}

func CoInitialize() error {
	r, _, err := procCoInitializeEx.Call(0, 2)
	if int32(r) < 0 {
		return err
	}
	return nil
}

func CoUninitialize() {
	ignoreSyscall3(procCoUninitialize.Call())
}

type VirtualDesktopManager struct {
	ptr uintptr
}

var (
	clsidVirtualDesktopManager = GUID{Data1: 0xaa509086, Data2: 0x5ca9, Data3: 0x4c25, Data4: [8]byte{0x8f, 0x95, 0x58, 0x9d, 0x3c, 0x07, 0xb4, 0x8a}}
	iidVirtualDesktopManager   = GUID{Data1: 0xa5cd92ff, Data2: 0x29be, Data3: 0x454c, Data4: [8]byte{0x8d, 0x04, 0xd8, 0x28, 0x79, 0xfb, 0x3f, 0x1b}}
)

func NewVirtualDesktopManager() (*VirtualDesktopManager, error) {
	var ptr uintptr
	r, _, err := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidVirtualDesktopManager)),
		0,
		CLSCTX_INPROC_SERVER,
		uintptr(unsafe.Pointer(&iidVirtualDesktopManager)),
		uintptr(unsafe.Pointer(&ptr)),
	)
	if int32(r) < 0 {
		return nil, err
	}
	return &VirtualDesktopManager{ptr: ptr}, nil
}

type virtualDesktopManagerVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	IsCurrent      uintptr
	GetDesktopID   uintptr
	MoveToDesktop  uintptr
}

func (m *VirtualDesktopManager) vtbl() *virtualDesktopManagerVtbl {
	if m == nil || m.ptr == 0 {
		return nil
	}
	return *(**virtualDesktopManagerVtbl)(unsafe.Pointer(m.ptr))
}

func (m *VirtualDesktopManager) Release() {
	if m == nil || m.ptr == 0 {
		return
	}
	vtbl := m.vtbl()
	if vtbl == nil || vtbl.Release == 0 {
		m.ptr = 0
		return
	}
	_, _, _ = syscall.SyscallN(vtbl.Release, m.ptr)
	m.ptr = 0
}

func (m *VirtualDesktopManager) IsWindowOnCurrentDesktop(hwnd HWND) (bool, error) {
	if m == nil || m.ptr == 0 {
		return true, nil
	}
	vtbl := m.vtbl()
	if vtbl == nil || vtbl.IsCurrent == 0 {
		return true, nil
	}
	var onCurrent int32
	r, _, err := syscall.SyscallN(vtbl.IsCurrent, m.ptr, uintptr(hwnd), uintptr(unsafe.Pointer(&onCurrent)))
	if int32(r) < 0 {
		return false, err
	}
	return onCurrent != 0, nil
}
