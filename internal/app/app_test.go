package app

import (
	"errors"
	"io"
	"log"
	"reflect"
	"testing"

	"better_alt_tab/internal/config"
	"better_alt_tab/internal/mru"
	"better_alt_tab/internal/session"
	"better_alt_tab/internal/ui"
	"better_alt_tab/internal/windows"
)

func TestTrayNotificationCodeUsesLowWord(t *testing.T) {
	got := trayNotificationCode(0x1234007B)
	if got != wmContextMenu {
		t.Fatalf("got %#x want %#x", got, wmContextMenu)
	}
}

func TestHandleCommandOpensSettings(t *testing.T) {
	a := newTestApp()
	calls := 0
	a.openSettings = func() error {
		calls++
		return nil
	}

	handled := a.handleCommand(ui.CommandOpenSettings)
	if !handled {
		t.Fatal("expected open settings command to be handled")
	}
	if calls != 1 {
		t.Fatalf("got open settings calls %d want 1", calls)
	}
}

func TestHandleCommandOpensConfigFile(t *testing.T) {
	a := newTestApp()
	calls := 0
	a.openConfigFile = func() error {
		calls++
		return nil
	}

	handled := a.handleCommand(ui.CommandOpenConfigFile)
	if !handled {
		t.Fatal("expected open config file command to be handled")
	}
	if calls != 1 {
		t.Fatalf("got open config file calls %d want 1", calls)
	}
}

func TestHandleCommandIgnoresUnknownCommand(t *testing.T) {
	a := newTestApp()

	if a.handleCommand(9999) {
		t.Fatal("expected unknown command to be ignored")
	}
}

func TestNormalizeCandidateIndex(t *testing.T) {
	tests := []struct {
		name  string
		start int
		count int
		want  int
	}{
		{name: "empty", start: 4, count: 0, want: 0},
		{name: "in range", start: 1, count: 3, want: 1},
		{name: "wraps forward", start: 4, count: 3, want: 1},
		{name: "wraps negative", start: -1, count: 3, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCandidateIndex(tt.start, tt.count); got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestActivateCandidateFromSkipsInvalidAndFailedTargets(t *testing.T) {
	activated := make([]windows.WindowID, 0, 2)
	a := newTestApp()
	a.session.Candidates = []windows.WindowID{10, 20, 30}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		if target == 10 {
			return errors.New("activate failed")
		}
		return nil
	}

	index, target, err := a.activateCandidateFrom(0, 0)
	if err != nil {
		t.Fatalf("activateCandidateFrom returned error: %v", err)
	}
	if index != 2 {
		t.Fatalf("got index %d want 2", index)
	}
	if target != 30 {
		t.Fatalf("got target %v want 30", target)
	}
	if !reflect.DeepEqual(activated, []windows.WindowID{10, 30}) {
		t.Fatalf("got activations %v want [10 30]", activated)
	}
}

func TestPreviewSelectionUpdatesIndexAndSnapshotTarget(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.lastSnapshot = windows.InventorySnapshot{
		ByID: map[windows.WindowID]windows.WindowInfo{
			10: {ID: 10},
			20: {ID: 20},
			30: {ID: 30},
		},
	}
	var activated windows.WindowID
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }
	a.activateTarget = func(target windows.WindowID) error {
		activated = target
		return nil
	}

	if !a.previewSelection() {
		t.Fatal("expected previewSelection to succeed")
	}
	if activated != 20 {
		t.Fatalf("got activated %v want 20", activated)
	}
	if a.session.SelectedIndex != 1 {
		t.Fatalf("got selected index %d want 1", a.session.SelectedIndex)
	}
	if _, ok := a.lastSnapshot.ByID[20]; !ok {
		t.Fatal("expected selected target to be remembered in snapshot")
	}
}

func TestOnTabPressedWithoutInstantPreviewOnlyUpdatesSelection(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onTabPressed()

	if !reflect.DeepEqual(activated, []windows.WindowID(nil)) {
		t.Fatalf("got activations %v want none", activated)
	}
	if a.session.SelectedIndex != 2 {
		t.Fatalf("got selected index %d want 2", a.session.SelectedIndex)
	}
}

func TestFinalizeSelectionFallsBackToNextActivatableTarget(t *testing.T) {
	a := newTestApp()
	a.mru = mru.New()
	a.session = session.SwitchSession{
		State:         session.StateCommitPending,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}
	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	if err := a.finalizeSelection(20); err != nil {
		t.Fatalf("finalizeSelection returned error: %v", err)
	}
	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
	got := a.mru.BuildCandidates(windows.InventorySnapshot{
		Order: []windows.WindowID{10, 20, 30},
	})
	want := []windows.WindowID{30, 10, 20}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got MRU order %v want %v", got, want)
	}
}

func TestCancelSessionRestoresStartingWindowWhenValid(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:          session.StateCycling,
		Candidates:     []windows.WindowID{10, 20},
		SelectedIndex:  1,
		StartedFrom:    10,
		OverlayVisible: true,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return target == 10 }
	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.cancelSession()

	if !reflect.DeepEqual(activated, []windows.WindowID{10}) {
		t.Fatalf("got activations %v want [10]", activated)
	}
	if a.session.State != session.StateIdle {
		t.Fatalf("got state %v want idle", a.session.State)
	}
	if a.session.OverlayVisible {
		t.Fatal("expected overlay hidden after cancel")
	}
}

func TestOnAltReleasedFinalizesSelectionAndReleasesModifiers(t *testing.T) {
	a := newTestApp()
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	releaseCalls := 0
	a.releaseModifiers = func() error {
		releaseCalls++
		return nil
	}

	a.onAltReleased()

	if releaseCalls != 1 {
		t.Fatalf("got release calls %d want 1", releaseCalls)
	}
	if a.session.State != session.StateIdle {
		t.Fatalf("got state %v want idle", a.session.State)
	}
}

func TestOnAltReleasedWithoutInstantPreviewActivatesSelectedTarget(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 2,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onAltReleased()

	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
}

func TestOnAltReleasedWithoutInstantPreviewFallsBackToNextValidTarget(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:         session.StateCycling,
		Candidates:    []windows.WindowID{10, 20, 30},
		SelectedIndex: 1,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool {
		return target != 20
	}

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.onAltReleased()

	if !reflect.DeepEqual(activated, []windows.WindowID{30}) {
		t.Fatalf("got activations %v want [30]", activated)
	}
}

func TestCancelSessionWithoutInstantPreviewDoesNotRestoreStartingWindow(t *testing.T) {
	a := newTestApp()
	a.cfg.InstantSwitchPreview = false
	a.session = session.SwitchSession{
		State:          session.StateCycling,
		Candidates:     []windows.WindowID{10, 20},
		SelectedIndex:  1,
		StartedFrom:    10,
		OverlayVisible: true,
	}
	a.isValidSwitchTarget = func(target windows.WindowID) bool { return true }

	var activated []windows.WindowID
	a.activateTarget = func(target windows.WindowID) error {
		activated = append(activated, target)
		return nil
	}

	a.cancelSession()

	if len(activated) != 0 {
		t.Fatalf("got activations %v want none", activated)
	}
}

func TestSaveSettingsConfigUpdatesRuntimeAndStartup(t *testing.T) {
	a := newTestApp()
	want := config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: false,
	}
	saved := config.Config{}
	startupCalls := 0
	a.saveConfig = func(cfg config.Config) error {
		saved = cfg
		return nil
	}
	a.syncStartup = func(enabled bool) error {
		startupCalls++
		if !enabled {
			t.Fatal("expected startup sync to receive enabled=true")
		}
		return nil
	}

	if err := a.saveSettingsConfig(want); err != nil {
		t.Fatalf("saveSettingsConfig returned error: %v", err)
	}
	if saved != want {
		t.Fatalf("got saved config %+v want %+v", saved, want)
	}
	if a.cfg != want {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, want)
	}
	if startupCalls != 1 {
		t.Fatalf("got startup calls %d want 1", startupCalls)
	}
}

func TestLoadCurrentConfigUpdatesRuntimeWithoutSyncingStartup(t *testing.T) {
	a := newTestApp()
	want := config.Config{
		ShowThumbnails:       false,
		LaunchOnStartup:      true,
		InstantSwitchPreview: true,
	}
	a.loadConfig = func() (config.Config, error) {
		return want, nil
	}
	startupCalls := 0
	a.syncStartup = func(bool) error {
		startupCalls++
		return nil
	}

	got, err := a.loadCurrentConfig()
	if err != nil {
		t.Fatalf("loadCurrentConfig returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got loaded config %+v want %+v", got, want)
	}
	if a.cfg != want {
		t.Fatalf("got runtime config %+v want %+v", a.cfg, want)
	}
	if startupCalls != 0 {
		t.Fatalf("got startup calls %d want 0", startupCalls)
	}
}

func newTestApp() *App {
	a := &App{
		logger:  log.New(io.Discard, "", 0),
		cfg:     config.Default(),
		mru:     mru.New(),
		overlay: ui.NewOverlay(0),
	}
	a.openSettings = func() error { return nil }
	a.openConfigFile = func() error { return nil }
	return a
}
