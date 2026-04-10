package windows

import "testing"

func TestDesktopManagerFailsClosedWithoutManager(t *testing.T) {
	manager := &DesktopManager{}
	if manager.IsWindowOnCurrentDesktop(1) {
		t.Fatal("expected missing desktop manager to exclude the window")
	}
}
