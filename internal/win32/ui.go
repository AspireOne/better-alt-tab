package win32

import (
	"fmt"
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
	ignoreSyscall3(procSetWindowPos.Call(
		uintptr(hwnd),
		0,
		uintptrFromInt32(x),
		uintptrFromInt32(y),
		uintptrFromInt32(width),
		uintptrFromInt32(height),
		flags,
	))
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

func InvalidateRectArea(hwnd HWND, rect RECT) {
	// #nosec G103 -- Win32 syscall boundary requires passing the RECT input pointer.
	ignoreSyscall3(procInvalidateRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)), 0))
}

func LoadDefaultApplicationIcon() HICON {
	r, _, _ := procLoadIconW.Call(0, IDI_APPLICATION)
	return HICON(r)
}

func GetWindowIcon(hwnd HWND) HICON {
	if hwnd == 0 {
		return 0
	}
	for _, iconType := range []uintptr{ICON_BIG, ICON_SMALL2, ICON_SMALL} {
		var result uintptr
		r, _, _ := procSendMessageTimeoutW.Call(
			uintptr(hwnd),
			WM_GETICON,
			iconType,
			0,
			SMTO_ABORTIFHUNG|SMTO_ERRORONEXIT,
			20,
			// #nosec G103 -- Win32 syscall boundary requires passing the result pointer.
			uintptr(unsafe.Pointer(&result)),
		)
		if r != 0 && result != 0 {
			return HICON(result)
		}
	}
	return 0
}

func GetClassIcon(hwnd HWND) HICON {
	if hwnd == 0 {
		return 0
	}
	for _, index := range []int32{GCLP_HICON, GCLP_HICONSM} {
		r, _, _ := procGetClassLongPtrW.Call(uintptr(hwnd), uintptrFromInt32(index))
		if r != 0 {
			return HICON(r)
		}
	}
	return 0
}

func GetShellIcon(path string) (HICON, bool) {
	if path == "" {
		return 0, false
	}
	var info SHFILEINFO
	r, _, _ := procSHGetFileInfoW.Call(
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 path pointer.
		uintptr(unsafe.Pointer(utf16Ptr(path))),
		0,
		// #nosec G103 -- Win32 syscall boundary requires passing the output struct pointer.
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
		SHGFI_ICON|SHGFI_LARGEICON,
	)
	if r == 0 || info.HIcon == 0 {
		return 0, false
	}
	return info.HIcon, true
}

func DestroyIcon(icon HICON) {
	if icon == 0 {
		return
	}
	ignoreSyscall3(procDestroyIcon.Call(uintptr(icon)))
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
	// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 message name.
	r, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(utf16Ptr(name))))
	return uint32FromRet(r)
}

func CreateSolidBrush(color uintptr) HBRUSH {
	r, _, _ := procCreateSolidBrush.Call(color)
	return HBRUSH(r)
}

func CreateCompatibleDC(hdc HDC) HDC {
	r, _, _ := procCreateCompatibleDC.Call(uintptr(hdc))
	return HDC(r)
}

func DeleteDC(hdc HDC) {
	ignoreSyscall3(procDeleteDC.Call(uintptr(hdc)))
}

func CreateCompatibleBitmap(hdc HDC, width, height int32) HBITMAP {
	r, _, _ := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptrFromInt32(width), uintptrFromInt32(height))
	return HBITMAP(r)
}

func SelectObject(hdc HDC, obj HGDIOBJ) HGDIOBJ {
	r, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return HGDIOBJ(r)
}

func DeleteObject(obj uintptr) {
	ignoreSyscall3(procDeleteObject.Call(obj))
}

func BeginPaint(hwnd HWND) (PAINTSTRUCT, HDC) {
	var ps PAINTSTRUCT
	// #nosec G103 -- Win32 syscall boundary requires passing the PAINTSTRUCT output pointer.
	r, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	return ps, HDC(r)
}

func EndPaint(hwnd HWND, ps *PAINTSTRUCT) {
	// #nosec G103 -- Win32 syscall boundary requires passing the PAINTSTRUCT input pointer.
	ignoreSyscall3(procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps))))
}

func FillRect(hdc HDC, rect *RECT, brush HBRUSH) {
	// #nosec G103 -- Win32 syscall boundary requires passing the RECT input pointer.
	ignoreSyscall3(procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush)))
}

func DrawLabel(hdc HDC, rect RECT, text string, color uintptr) {
	if hdc == 0 || text == "" || rect.Right <= rect.Left || rect.Bottom <= rect.Top {
		return
	}
	oldFont := SelectObject(hdc, HGDIOBJ(defaultGUIFont()))
	defer SelectObject(hdc, oldFont)
	ignoreSyscall3(procSetBkMode.Call(uintptr(hdc), TRANSPARENT))
	ignoreSyscall3(procSetTextColor.Call(uintptr(hdc), color))
	flags := uintptr(DT_CENTER | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS | DT_NOPREFIX)
	ignoreSyscall3(procDrawTextW.Call(
		uintptr(hdc),
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 text buffer.
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		^uintptr(0),
		// #nosec G103 -- Win32 syscall boundary requires passing the RECT input pointer.
		uintptr(unsafe.Pointer(&rect)),
		flags,
	))
}

func defaultGUIFont() HFONT {
	r, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
	return HFONT(r)
}

func DrawIconInRect(hdc HDC, rect RECT, icon HICON) {
	ignoreSyscall3(procDrawIconEx.Call(
		uintptr(hdc),
		uintptrFromInt32(rect.Left),
		uintptrFromInt32(rect.Top),
		uintptr(icon),
		uintptrFromInt32(rect.Right-rect.Left),
		uintptrFromInt32(rect.Bottom-rect.Top),
		0,
		0,
		3,
	))
}

func SetStretchBltMode(hdc HDC, mode int32) {
	ignoreSyscall3(procSetStretchBltMode.Call(uintptr(hdc), uintptrFromInt32(mode)))
}

func BitBlt(dst HDC, x, y, width, height int32, src HDC, srcX, srcY int32, rop uint32) bool {
	r, _, _ := procBitBlt.Call(
		uintptr(dst),
		uintptrFromInt32(x),
		uintptrFromInt32(y),
		uintptrFromInt32(width),
		uintptrFromInt32(height),
		uintptr(src),
		uintptrFromInt32(srcX),
		uintptrFromInt32(srcY),
		uintptr(rop),
	)
	return r != 0
}

func StretchBlt(dst HDC, x, y, width, height int32, src HDC, srcX, srcY, srcWidth, srcHeight int32, rop uint32) bool {
	r, _, _ := procStretchBlt.Call(
		uintptr(dst),
		uintptrFromInt32(x),
		uintptrFromInt32(y),
		uintptrFromInt32(width),
		uintptrFromInt32(height),
		uintptr(src),
		uintptrFromInt32(srcX),
		uintptrFromInt32(srcY),
		uintptrFromInt32(srcWidth),
		uintptrFromInt32(srcHeight),
		uintptr(rop),
	)
	return r != 0
}

func PrintWindow(hwnd HWND, hdc HDC, flags uint32) bool {
	r, _, _ := procPrintWindow.Call(uintptr(hwnd), uintptr(hdc), uintptr(flags))
	return r != 0
}

func DrawBitmapInRect(hdc HDC, rect RECT, bitmap HBITMAP, srcWidth, srcHeight int32) bool {
	if hdc == 0 || bitmap == 0 || srcWidth <= 0 || srcHeight <= 0 {
		return false
	}
	memDC := CreateCompatibleDC(hdc)
	if memDC == 0 {
		return false
	}
	defer DeleteDC(memDC)
	old := SelectObject(memDC, HGDIOBJ(bitmap))
	defer SelectObject(memDC, old)
	SetStretchBltMode(hdc, HALFTONE)
	return StretchBlt(
		hdc,
		rect.Left,
		rect.Top,
		rect.Right-rect.Left,
		rect.Bottom-rect.Top,
		memDC,
		0,
		0,
		srcWidth,
		srcHeight,
		SRCCOPY,
	)
}

func MonitorFromWindow(hwnd HWND) HMONITOR {
	r, _, _ := procMonitorFromWindow.Call(uintptr(hwnd), MONITOR_DEFAULTTONEAREST)
	return HMONITOR(r)
}

func GetMonitorRect(monitor HMONITOR) RECT {
	info := MONITORINFO{CbSize: uint32(unsafe.Sizeof(MONITORINFO{}))}
	// #nosec G103 -- Win32 syscall boundary requires passing the MONITORINFO output pointer.
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
		// #nosec G103 -- Win32 syscall boundary requires passing the INPUT array pointer.
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
	// #nosec G103 -- Win32 syscall boundary requires passing the destination UTF-16 buffer.
	r, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:r])
}

func IsWindowCloaked(hwnd HWND) bool {
	var cloaked uint32
	// #nosec G103 -- Win32 syscall boundary requires passing the cloaked-state output pointer.
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
	// #nosec G103 -- Win32 syscall boundary requires passing the destination UTF-16 buffer.
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
	// #nosec G103 -- Win32 syscall boundary requires passing the NOTIFYICONDATA input pointer.
	r, _, err := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&data)))
	if r == 0 {
		return err
	}
	data.UTimeoutOrVersion = NOTIFYICON_VERSION_4
	// #nosec G103 -- Win32 syscall boundary requires passing the NOTIFYICONDATA input pointer.
	ignoreSyscall3(procShellNotifyIconW.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&data))))
	return nil
}

func DeleteTrayIcon(hwnd HWND) error {
	var data NOTIFYICONDATA
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.HWnd = hwnd
	data.UID = 1
	// #nosec G103 -- Win32 syscall boundary requires passing the NOTIFYICONDATA input pointer.
	r, _, err := procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&data)))
	if r == 0 {
		return err
	}
	return nil
}

func ShowTrayMenu(hwnd HWND, settingsCommandID, openConfigCommandID, reloadThemeCommandID, exitCommandID uint32) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer func() {
		ignoreSyscall3(procDestroyMenu.Call(menu))
	}()
	if settingsCommandID != 0 {
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 menu label.
		ignoreSyscall3(procAppendMenuW.Call(menu, MF_STRING, uintptr(settingsCommandID), uintptr(unsafe.Pointer(utf16Ptr("Settings")))))
	}
	// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 menu label.
	ignoreSyscall3(procAppendMenuW.Call(menu, MF_STRING, uintptr(openConfigCommandID), uintptr(unsafe.Pointer(utf16Ptr("Open Config File")))))
	if reloadThemeCommandID != 0 {
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 menu label.
		ignoreSyscall3(procAppendMenuW.Call(menu, MF_STRING, uintptr(reloadThemeCommandID), uintptr(unsafe.Pointer(utf16Ptr("Reload Theme")))))
	}
	// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 menu label.
	ignoreSyscall3(procAppendMenuW.Call(menu, MF_STRING, uintptr(exitCommandID), uintptr(unsafe.Pointer(utf16Ptr("Close")))))
	var pt POINT
	// #nosec G103 -- Win32 syscall boundary requires passing the POINT output pointer.
	ignoreSyscall3(procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))))
	cmd, _, _ := procTrackPopupMenu.Call(menu, TPM_RETURNCMD|TPM_NONOTIFY, uintptrFromInt32(pt.X), uintptrFromInt32(pt.Y), 0, uintptr(hwnd), 0)
	if cmd != 0 {
		PostMessage(hwnd, WM_COMMAND, cmd, 0)
	}
}

func OpenPath(path string) error {
	r, _, _ := procShellExecuteW.Call(
		0,
		// #nosec G103 -- Win32 syscall boundary requires passing a UTF-16 verb pointer.
		uintptr(unsafe.Pointer(utf16Ptr("open"))),
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 path pointer.
		uintptr(unsafe.Pointer(utf16Ptr(path))),
		0,
		0,
		SW_SHOW,
	)
	if r <= 32 {
		return fmt.Errorf("ShellExecuteW failed for %q: code %d", path, r)
	}
	return nil
}

func SetCheckboxChecked(hwnd HWND, checked bool) {
	if hwnd == 0 {
		return
	}
	value := uintptr(0)
	if checked {
		value = BST_CHECKED
	}
	SendMessage(hwnd, BM_SETCHECK, value, 0)
}

func CheckboxChecked(hwnd HWND) bool {
	if hwnd == 0 {
		return false
	}
	return SendMessage(hwnd, BM_GETCHECK, 0, 0) == BST_CHECKED
}

func ShowErrorMessage(hwnd HWND, title, text string) {
	ignoreSyscall3(procMessageBoxW.Call(
		uintptr(hwnd),
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 message text pointer.
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		// #nosec G103 -- Win32 syscall boundary requires passing the UTF-16 message title pointer.
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		MB_OK|MB_ICONERROR,
	))
}

func CoInitialize() error {
	r, _, err := procCoInitializeEx.Call(0, 2)
	if hresultFailed(r) {
		return err
	}
	return nil
}

func CoUninitialize() {
	ignoreSyscall3(procCoUninitialize.Call())
}

type VirtualDesktopManager struct {
	ptr unsafe.Pointer
}

var (
	clsidVirtualDesktopManager = GUID{Data1: 0xaa509086, Data2: 0x5ca9, Data3: 0x4c25, Data4: [8]byte{0x8f, 0x95, 0x58, 0x9d, 0x3c, 0x07, 0xb4, 0x8a}}
	iidVirtualDesktopManager   = GUID{Data1: 0xa5cd92ff, Data2: 0x29be, Data3: 0x454c, Data4: [8]byte{0x8d, 0x04, 0xd8, 0x28, 0x79, 0xfb, 0x3f, 0x1b}}
)

func NewVirtualDesktopManager() (*VirtualDesktopManager, error) {
	var ptr unsafe.Pointer
	r, _, err := procCoCreateInstance.Call(
		// #nosec G103 -- COM activation requires passing CLSID/IID/output pointers.
		uintptr(unsafe.Pointer(&clsidVirtualDesktopManager)),
		0,
		CLSCTX_INPROC_SERVER,
		// #nosec G103 -- COM activation requires passing CLSID/IID/output pointers.
		uintptr(unsafe.Pointer(&iidVirtualDesktopManager)),
		// #nosec G103 -- COM activation requires passing CLSID/IID/output pointers.
		uintptr(unsafe.Pointer(&ptr)),
	)
	if hresultFailed(r) {
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
	if m == nil || m.ptr == nil {
		return nil
	}
	return *(**virtualDesktopManagerVtbl)(m.ptr)
}

func (m *VirtualDesktopManager) Release() {
	if m == nil || m.ptr == nil {
		return
	}
	vtbl := m.vtbl()
	if vtbl == nil || vtbl.Release == 0 {
		m.ptr = nil
		return
	}
	_, _, _ = syscall.SyscallN(vtbl.Release, uintptr(m.ptr))
	m.ptr = nil
}

func (m *VirtualDesktopManager) IsWindowOnCurrentDesktop(hwnd HWND) (bool, error) {
	if m == nil || m.ptr == nil {
		return true, nil
	}
	vtbl := m.vtbl()
	if vtbl == nil || vtbl.IsCurrent == 0 {
		return true, nil
	}
	var onCurrent int32
	// #nosec G103 -- COM call requires passing the BOOL output pointer.
	r, _, err := syscall.SyscallN(vtbl.IsCurrent, uintptr(m.ptr), uintptr(hwnd), uintptr(unsafe.Pointer(&onCurrent)))
	if hresultFailed(r) {
		return false, err
	}
	return onCurrent != 0, nil
}
