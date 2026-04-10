package session

import (
	"testing"

	"quick_app_switcher/internal/windows"
)

func TestStartSelectsSecondCandidate(t *testing.T) {
	var s SwitchSession
	ok := s.Start([]windows.WindowID{1, 2, 3}, 1)
	if !ok {
		t.Fatal("expected session start")
	}
	current, ok := s.Current()
	if !ok {
		t.Fatal("expected current candidate")
	}
	if current != 2 {
		t.Fatalf("got %v want 2", current)
	}
}

func TestAdvanceWraps(t *testing.T) {
	var s SwitchSession
	s.Start([]windows.WindowID{1, 2, 3}, 1)
	s.Advance()
	s.Advance()
	current, _ := s.Current()
	if current != 1 {
		t.Fatalf("got %v want 1", current)
	}
}

func TestStartRejectsSingleCandidate(t *testing.T) {
	var s SwitchSession
	if s.Start([]windows.WindowID{1}, 1) {
		t.Fatal("expected start to fail")
	}
	if s.State != StateIdle {
		t.Fatalf("got %v want idle", s.State)
	}
}
