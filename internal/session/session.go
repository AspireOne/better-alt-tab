package session

import "better_alt_tab/internal/windows"

type State uint8

const (
	StateIdle State = iota
	StateCycling
	StateCommitPending
	StateCancelled
)

type SwitchSession struct {
	State          State
	Candidates     []windows.WindowID
	SelectedIndex  int
	StartedFrom    windows.WindowID
	OverlayVisible bool
}

func (s *SwitchSession) Start(candidates []windows.WindowID, startedFrom windows.WindowID) bool {
	s.Reset()
	if len(candidates) < 2 {
		return false
	}
	s.State = StateCycling
	s.Candidates = append(s.Candidates, candidates...)
	s.SelectedIndex = 1
	s.StartedFrom = startedFrom
	s.OverlayVisible = true
	return true
}

func (s *SwitchSession) Advance() bool {
	if s.State != StateCycling || len(s.Candidates) < 2 {
		return false
	}
	s.SelectedIndex = (s.SelectedIndex + 1) % len(s.Candidates)
	return true
}

func (s *SwitchSession) Current() (windows.WindowID, bool) {
	if len(s.Candidates) == 0 || s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Candidates) {
		return 0, false
	}
	return s.Candidates[s.SelectedIndex], true
}

func (s *SwitchSession) BeginCommit() (windows.WindowID, bool) {
	current, ok := s.Current()
	if !ok || s.State != StateCycling {
		return 0, false
	}
	s.State = StateCommitPending
	s.OverlayVisible = false
	return current, true
}

func (s *SwitchSession) Cancel() {
	s.State = StateCancelled
	s.OverlayVisible = false
	s.Candidates = s.Candidates[:0]
	s.SelectedIndex = 0
	s.StartedFrom = 0
}

func (s *SwitchSession) Reset() {
	s.State = StateIdle
	s.OverlayVisible = false
	s.Candidates = s.Candidates[:0]
	s.SelectedIndex = 0
	s.StartedFrom = 0
}
