package windows

import (
	"path/filepath"
	"strings"

	"better_alt_tab/internal/win32"
)

type WindowID uintptr

func (id WindowID) HWND() win32.HWND {
	return win32.HWND(id)
}

type WindowInfo struct {
	ID               WindowID
	Title            string
	ProcessID        uint32
	ExecutablePath   string
	Visible          bool
	Minimized        bool
	Cloaked          bool
	OnCurrentDesktop bool
	Style            uintptr
	ExStyle          uintptr
	Owner            WindowID
	RootOwner        WindowID
	LastActivePopup  WindowID
	ClassName        string
	IsAppWindow      bool
}

type InventorySnapshot struct {
	Order []WindowID
	ByID  map[WindowID]WindowInfo
}

func (w WindowInfo) AppDisplayName() string {
	if w.ExecutablePath != "" {
		name := strings.TrimSuffix(filepath.Base(w.ExecutablePath), filepath.Ext(w.ExecutablePath))
		if name != "" && name != "." && name != string(filepath.Separator) {
			return name
		}
	}
	if w.Title != "" {
		return w.Title
	}
	return w.ClassName
}

func (s InventorySnapshot) Set() map[WindowID]struct{} {
	out := make(map[WindowID]struct{}, len(s.Order))
	for _, id := range s.Order {
		out[id] = struct{}{}
	}
	return out
}
