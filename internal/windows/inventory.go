package windows

import "quick_app_switcher/internal/win32"

type Inventory struct {
	filter  Filter
	desktop *DesktopManager
}

func NewInventory(excluded []win32.HWND, desktop *DesktopManager) *Inventory {
	excludedSet := make(map[WindowID]struct{}, len(excluded))
	for _, hwnd := range excluded {
		excludedSet[WindowID(hwnd)] = struct{}{}
	}
	return &Inventory{
		filter:  Filter{Excluded: excludedSet},
		desktop: desktop,
	}
}

func (i *Inventory) Snapshot() (InventorySnapshot, error) {
	raw := make(map[WindowID]WindowInfo)
	err := win32.EnumWindows(func(hwnd win32.HWND) bool {
		info := i.inspectWindow(hwnd)
		if info.ID != 0 {
			raw[info.ID] = info
		}
		return true
	})
	if err != nil {
		return InventorySnapshot{}, err
	}

	byID := make(map[WindowID]WindowInfo, len(raw))
	order := make([]WindowID, 0, len(raw))
	for id, info := range raw {
		info.IsAppWindow = i.filter.Eligible(info, raw)
		if !info.IsAppWindow {
			continue
		}
		byID[id] = info
		order = append(order, id)
	}
	return InventorySnapshot{Order: order, ByID: byID}, nil
}

func (i *Inventory) IsValidSwitchable(id WindowID) bool {
	snapshot, err := i.Snapshot()
	if err != nil {
		return false
	}
	_, ok := snapshot.ByID[id]
	return ok
}

func (i *Inventory) inspectWindow(hwnd win32.HWND) WindowInfo {
	if !win32.IsWindow(hwnd) {
		return WindowInfo{}
	}
	info := WindowInfo{
		ID:               WindowID(hwnd),
		Title:            win32.GetWindowText(hwnd),
		ProcessID:        win32.GetWindowProcessID(hwnd),
		ExecutablePath:   win32.GetWindowProcessPath(hwnd),
		Visible:          win32.IsWindowVisible(hwnd),
		Minimized:        win32.IsIconic(hwnd),
		Cloaked:          win32.IsWindowCloaked(hwnd),
		OnCurrentDesktop: i.desktop == nil || i.desktop.IsWindowOnCurrentDesktop(hwnd),
		Style:            win32.GetWindowStyle(hwnd),
		ExStyle:          win32.GetWindowExStyle(hwnd),
		Owner:            WindowID(win32.GetWindow(hwnd, win32.GW_OWNER)),
		RootOwner:        WindowID(win32.GetAncestor(hwnd, win32.GA_ROOTOWNER)),
		LastActivePopup:  WindowID(win32.GetLastActivePopup(hwnd)),
		ClassName:        win32.GetClassName(hwnd),
	}
	return info
}
