package windows

import "quick_app_switcher/internal/win32"

type Filter struct {
	Excluded map[WindowID]struct{}
}

func (f Filter) Eligible(info WindowInfo, all map[WindowID]WindowInfo) bool {
	if !f.baseEligible(info) {
		return false
	}
	return f.representativeFor(info, all) == info.ID
}

func (f Filter) baseEligible(info WindowInfo) bool {
	if info.ID == 0 || !info.Visible {
		return false
	}
	if _, ok := f.Excluded[info.ID]; ok {
		return false
	}
	if info.ExStyle&win32.WS_EX_TOOLWINDOW != 0 {
		return false
	}
	if info.Cloaked || !info.OnCurrentDesktop {
		return false
	}
	if info.ClassName == "Progman" || info.ClassName == "WorkerW" || info.ClassName == "Shell_TrayWnd" {
		return false
	}
	return true
}

func (f Filter) representativeFor(info WindowInfo, all map[WindowID]WindowInfo) WindowID {
	rootID := info.RootOwner
	if rootID == 0 {
		rootID = info.ID
	}

	rootInfo, ok := all[rootID]
	if !ok {
		rootInfo = info
		rootID = info.ID
	}

	targetID := rootID
	if popupID := rootInfo.LastActivePopup; popupID != 0 && popupID != rootID {
		if popup, ok := all[popupID]; ok && f.baseEligible(popup) {
			targetID = popupID
		}
	}

	targetInfo, ok := all[targetID]
	if !ok || !f.baseEligible(targetInfo) {
		return 0
	}
	return targetID
}
