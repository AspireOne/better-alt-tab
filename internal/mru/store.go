package mru

import "quick_app_switcher/internal/windows"

type Store struct {
	order []windows.WindowID
	set   map[windows.WindowID]struct{}
}

func New() *Store {
	return &Store{set: make(map[windows.WindowID]struct{})}
}

func (s *Store) MoveToFront(id windows.WindowID) {
	if id == 0 {
		return
	}
	s.Remove(id)
	s.order = append([]windows.WindowID{id}, s.order...)
	s.set[id] = struct{}{}
}

func (s *Store) Remove(id windows.WindowID) {
	if id == 0 || len(s.order) == 0 {
		return
	}
	for i, existing := range s.order {
		if existing != id {
			continue
		}
		copy(s.order[i:], s.order[i+1:])
		s.order = s.order[:len(s.order)-1]
		break
	}
	delete(s.set, id)
}

func (s *Store) BuildCandidates(snapshot windows.InventorySnapshot) []windows.WindowID {
	seen := make(map[windows.WindowID]struct{}, len(snapshot.Order))
	out := make([]windows.WindowID, 0, len(snapshot.Order))
	valid := snapshot.Set()

	filtered := s.order[:0]
	for _, id := range s.order {
		if _, ok := valid[id]; !ok {
			delete(s.set, id)
			continue
		}
		filtered = append(filtered, id)
		out = append(out, id)
		seen[id] = struct{}{}
	}
	s.order = filtered

	for _, id := range snapshot.Order {
		if _, ok := seen[id]; ok {
			continue
		}
		out = append(out, id)
		seen[id] = struct{}{}
	}
	return out
}

func (s *Store) Seed(ids []windows.WindowID) {
	for i := len(ids) - 1; i >= 0; i-- {
		s.MoveToFront(ids[i])
	}
}
