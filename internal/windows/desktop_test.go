package windows

import "testing"

func TestDesktopManagerDegradesOpenWithoutManager(t *testing.T) {
	manager := &DesktopManager{}
	if !manager.IsWindowOnCurrentDesktop(1) {
		t.Fatal("expected missing desktop manager to include the window")
	}
}
