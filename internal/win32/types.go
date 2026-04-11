package win32

import (
	"syscall"
	"unsafe"
)

type (
	Handle    uintptr
	HWND      uintptr
	HICON     uintptr
	HCURSOR   uintptr
	HBRUSH    uintptr
	HDC       uintptr
	HBITMAP   uintptr
	HGDIOBJ   uintptr
	HMONITOR  uintptr
	HINSTANCE uintptr
	HMENU     uintptr
)

type POINT struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type MSG struct {
	HWnd     HWND
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

type PAINTSTRUCT struct {
	Hdc         HDC
	Erase       int32
	Paint       RECT
	Restore     int32
	IncUpdate   int32
	RGBReserved [32]byte
}

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   HINSTANCE
	Icon       HICON
	Cursor     HCURSOR
	Background HBRUSH
	MenuName   *uint16
	ClassName  *uint16
	IconSm     HICON
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

type KBDLLHOOKSTRUCT struct {
	VKCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type NOTIFYICONDATA struct {
	CbSize            uint32
	HWnd              HWND
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             HICON
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          GUID
	HBalloonIcon      HICON
}

type MOUSEINPUT struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type HARDWAREINPUT struct {
	UMsg    uint32
	WParamL uint16
	WParamH uint16
}

type INPUT struct {
	Type      uint32
	_         uint32
	Anonymous [32]byte
}

func NewKeyboardInput(vk uint16, scan uint16, flags uint32) INPUT {
	var input INPUT
	input.Type = INPUT_KEYBOARD
	kb := (*KEYBDINPUT)(unsafe.Pointer(&input.Anonymous[0]))
	kb.WVk = vk
	kb.WScan = scan
	kb.DwFlags = flags
	return input
}

const (
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	WS_POPUP         = 0x80000000
	WS_EX_TOPMOST    = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_LAYERED    = 0x00080000
	WS_EX_NOACTIVATE = 0x08000000
	WS_EX_APPWINDOW  = 0x00040000

	SW_HIDE    = 0
	SW_SHOW    = 5
	SW_RESTORE = 9

	SWP_NOACTIVATE = 0x0010
	SWP_SHOWWINDOW = 0x0040
	SWP_HIDEWINDOW = 0x0080

	LWA_ALPHA = 0x00000002

	WM_DESTROY   = 0x0002
	WM_PAINT     = 0x000F
	WM_COMMAND   = 0x0111
	WM_QUIT      = 0x0012
	WM_APP       = 0x8000
	WM_RBUTTONUP = 0x0205
	WM_LBUTTONUP = 0x0202

	WH_KEYBOARD_LL = 13
	LLKHF_UP       = 0x0080

	VK_TAB    = 0x09
	VK_ESCAPE = 0x1B
	VK_MENU   = 0x12
	VK_LMENU  = 0xA4
	VK_RMENU  = 0xA5

	GA_ROOTOWNER = 3
	GW_OWNER     = 4

	EVENT_SYSTEM_FOREGROUND = 0x0003
	WINEVENT_OUTOFCONTEXT   = 0x0000

	GWL_STYLE   = -16
	GWL_EXSTYLE = -20

	MONITOR_DEFAULTTONEAREST = 2
	DWMWA_CLOAKED            = 14

	NIF_MESSAGE          = 0x00000001
	NIF_ICON             = 0x00000002
	NIF_TIP              = 0x00000004
	NIM_ADD              = 0x00000000
	NIM_DELETE           = 0x00000002
	NIM_SETVERSION       = 0x00000004
	NOTIFYICON_VERSION_4 = 4

	TPM_RETURNCMD = 0x0100
	TPM_NONOTIFY  = 0x0080
	MF_STRING     = 0x0000

	IDI_APPLICATION = 32512

	INPUT_KEYBOARD  = 1
	KEYEVENTF_KEYUP = 0x0002

	CLSCTX_INPROC_SERVER = 0x1

	MAX_PATH = syscall.MAX_PATH

	SRCCOPY              = 0x00CC0020
	CAPTUREBLT           = 0x40000000
	HALFTONE             = 4
	PW_RENDERFULLCONTENT = 0x00000002
)
