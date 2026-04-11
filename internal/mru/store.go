package mru

import "better_alt_tab/internal/windows"

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
	if len(s.order) > 0 && s.order[0] == id {
		return
	}
	if _, ok := s.set[id]; ok {
		for i := 1; i < len(s.order); i++ {
			if s.order[i] != id {
				continue
			}
			copy(s.order[1:i+1], s.order[:i])
			s.order[0] = id
			return
		}
	}
	s.order = append(s.order, 0)
	copy(s.order[1:], s.order[:len(s.order)-1])
	s.order[0] = id
	s.set[id] = struct{}{}
}

func (s *Store) Remove(id windows.WindowID) {
	if id == 0 || len(s.order) == 0 {
		return
	}
	if _, ok := s.set[id]; !ok {
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
