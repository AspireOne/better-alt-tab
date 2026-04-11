package windows

import (
	"testing"

	"quick_app_switcher/internal/win32"
)

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

func TestEligibleTargetMatchesSnapshotEligibilityForLastActivePopup(t *testing.T) {
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
	all := map[WindowID]WindowInfo{
		root.ID:  root,
		popup.ID: popup,
	}

	if got, want := filter.EligibleTarget(popup, root, popup), filter.Eligible(popup, all); got != want {
		t.Fatalf("EligibleTarget(popup) = %v, want %v", got, want)
	}
	if got, want := filter.EligibleTarget(root, root, popup), filter.Eligible(root, all); got != want {
		t.Fatalf("EligibleTarget(root) = %v, want %v", got, want)
	}
}

func TestEligibleTargetMatchesSnapshotEligibilityWhenRootOwnsSwitchTarget(t *testing.T) {
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

	if got, want := filter.EligibleTarget(root, root, WindowInfo{}), filter.Eligible(root, all); got != want {
		t.Fatalf("EligibleTarget(root) = %v, want %v", got, want)
	}
	if got, want := filter.EligibleTarget(popup, root, WindowInfo{}), filter.Eligible(popup, all); got != want {
		t.Fatalf("EligibleTarget(popup) = %v, want %v", got, want)
	}
}

func TestEligibleRejectsToolWindow(t *testing.T) {
	filter := Filter{}
	info := WindowInfo{
		ID:               10,
		Visible:          true,
		OnCurrentDesktop: true,
		ExStyle:          win32.WS_EX_TOOLWINDOW,
		ClassName:        "Settings",
	}

	if filter.Eligible(info, map[WindowID]WindowInfo{info.ID: info}) {
		t.Fatal("expected tool window to be excluded from snapshot eligibility")
	}
	if filter.EligibleTarget(info, info, WindowInfo{}) {
		t.Fatal("expected tool window to be excluded from target eligibility")
	}
}
