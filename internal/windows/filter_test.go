package windows

import "testing"

func TestEligiblePrefersVisibleLastActivePopup(t *testing.T) {
	filter := Filter{}
	root := WindowInfo{
		ID:               10,
		RootOwner:        10,
		LastActivePopup:  20,
		OnCurrentDesktop: true,
		ClassName:        "Root",
	}
	popup := WindowInfo{
		ID:               20,
		Owner:            10,
		RootOwner:        10,
		LastActivePopup:  20,
		OnCurrentDesktop: true,
		Visible:          true,
		ClassName:        "Dialog",
	}
	root.Visible = false
	all := map[WindowID]WindowInfo{
		root.ID:  root,
		popup.ID: popup,
	}

	if filter.Eligible(root, all) {
		t.Fatal("expected hidden root owner to be excluded")
	}
	if !filter.Eligible(popup, all) {
		t.Fatal("expected visible last-active popup to be eligible")
	}
}

func TestEligibleRejectsOwnedPopupWhenRootOwnsSwitchTarget(t *testing.T) {
	filter := Filter{}
	root := WindowInfo{
		ID:               10,
		RootOwner:        10,
		LastActivePopup:  10,
		OnCurrentDesktop: true,
		Visible:          true,
		ClassName:        "Root",
	}
	popup := WindowInfo{
		ID:               20,
		Owner:            10,
		RootOwner:        10,
		LastActivePopup:  20,
		OnCurrentDesktop: true,
		Visible:          true,
		ClassName:        "Dialog",
	}
	all := map[WindowID]WindowInfo{
		root.ID:  root,
		popup.ID: popup,
	}

	if !filter.Eligible(root, all) {
		t.Fatal("expected root owner to remain eligible")
	}
	if filter.Eligible(popup, all) {
		t.Fatal("expected owned popup to be excluded when root owner is the switch target")
	}
}
