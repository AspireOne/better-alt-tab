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

func TestBeginCommitTransitionsWithoutChangingSelection(t *testing.T) {
	var s SwitchSession
	if !s.Start([]windows.WindowID{10, 20, 30}, 10) {
		t.Fatal("expected session start")
	}
	s.Advance()

	currentBefore, ok := s.Current()
	if !ok {
		t.Fatal("expected current candidate before commit")
	}

	currentAfter, ok := s.BeginCommit()
	if !ok {
		t.Fatal("expected commit to begin")
	}
	if currentAfter != currentBefore {
		t.Fatalf("got %v want %v", currentAfter, currentBefore)
	}
	if s.State != StateCommitPending {
		t.Fatalf("got %v want commit pending", s.State)
	}
	if s.OverlayVisible {
		t.Fatal("expected overlay hidden while commit is pending")
	}
}

func TestCancelClearsSessionState(t *testing.T) {
	var s SwitchSession
	if !s.Start([]windows.WindowID{10, 20, 30}, 10) {
		t.Fatal("expected session start")
	}

	s.Cancel()

	if s.State != StateCancelled {
		t.Fatalf("got %v want cancelled", s.State)
	}
	if s.OverlayVisible {
		t.Fatal("expected overlay hidden after cancel")
	}
	if len(s.Candidates) != 0 {
		t.Fatalf("got %d candidates want 0", len(s.Candidates))
	}
	if s.StartedFrom != 0 {
		t.Fatalf("got started from %v want 0", s.StartedFrom)
	}
}
