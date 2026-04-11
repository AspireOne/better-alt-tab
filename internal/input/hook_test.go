package input

import (
	"testing"

	"quick_app_switcher/internal/win32"
)

func TestHandleKeyAltReleaseEndsOwnedSessionWithoutSuppressingRelease(t *testing.T) {
	h := &Hook{
		altDown:      true,
		tabDown:      true,
		ownedSession: true,
	}

	decision := h.handleKey(win32.VK_MENU, win32.LLKHF_UP)

	if !decision.suppress {
		t.Fatal("expected Alt release to be suppressed for owned session")
	}
	if !decision.postAltUp {
		t.Fatal("expected Alt release to notify the app")
	}
	if h.altDown {
		t.Fatal("expected Alt state to be cleared")
	}
	if h.tabDown {
		t.Fatal("expected Tab state to be cleared on Alt release")
	}
	if h.ownedSession {
		t.Fatal("expected owned session to end on Alt release")
	}
}

func TestHandleKeyTabDownWhileAltHeldStartsOwnedSessionAndSuppressesTab(t *testing.T) {
	h := &Hook{altDown: true}

	decision := h.handleKey(win32.VK_TAB, 0)

	if !decision.suppress {
		t.Fatal("expected Tab down to be suppressed while Alt is held")
	}
	if !decision.postTab {
		t.Fatal("expected Tab down to notify the app")
	}
	if !h.tabDown {
		t.Fatal("expected Tab state to be set")
	}
	if !h.ownedSession {
		t.Fatal("expected session ownership to start")
	}
}

func TestHandleKeyInjectedAltEventsAreIgnored(t *testing.T) {
	h := &Hook{
		altDown:      true,
		tabDown:      true,
		ownedSession: true,
	}

	decision := h.handleKey(win32.VK_MENU, win32.LLKHF_INJECTED|win32.LLKHF_UP)

	if decision != (keyDecision{}) {
		t.Fatalf("got decision %+v want zero decision", decision)
	}
	if !h.altDown {
		t.Fatal("expected injected event to leave Alt state unchanged")
	}
	if !h.tabDown {
		t.Fatal("expected injected event to leave Tab state unchanged")
	}
	if !h.ownedSession {
		t.Fatal("expected injected event to leave ownership unchanged")
	}
}
