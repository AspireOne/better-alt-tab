package mru

import (
	"reflect"
	"testing"

	"quick_app_switcher/internal/windows"
)

func TestBuildCandidatesMergesMRUAndSnapshot(t *testing.T) {
	store := New()
	store.Seed([]windows.WindowID{3, 2, 1})
	got := store.BuildCandidates(windows.InventorySnapshot{
		Order: []windows.WindowID{2, 4, 1},
	})
	want := []windows.WindowID{2, 1, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMoveToFrontDeduplicates(t *testing.T) {
	store := New()
	store.Seed([]windows.WindowID{1, 2, 3})
	store.MoveToFront(2)
	got := store.BuildCandidates(windows.InventorySnapshot{
		Order: []windows.WindowID{1, 2, 3},
	})
	want := []windows.WindowID{2, 1, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
