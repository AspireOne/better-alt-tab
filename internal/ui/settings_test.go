package ui

import "testing"

func TestSettingsWindowHandleCommandDispatchesActions(t *testing.T) {
	w := NewSettingsWindow()

	saveCalls := 0
	handled, err := w.HandleCommand(controlSave, func() error {
		saveCalls++
		return nil
	}, nil)
	if err != nil {
		t.Fatalf("HandleCommand(save) returned error: %v", err)
	}
	if !handled {
		t.Fatal("expected save command to be handled")
	}
	if saveCalls != 1 {
		t.Fatalf("got save calls %d want 1", saveCalls)
	}

	cancelCalls := 0
	handled, err = w.HandleCommand(controlCancel, nil, func() {
		cancelCalls++
	})
	if err != nil {
		t.Fatalf("HandleCommand(cancel) returned error: %v", err)
	}
	if !handled {
		t.Fatal("expected cancel command to be handled")
	}
	if cancelCalls != 1 {
		t.Fatalf("got cancel calls %d want 1", cancelCalls)
	}

	handled, err = w.HandleCommand(9999, nil, nil)
	if err != nil {
		t.Fatalf("HandleCommand(unknown) returned error: %v", err)
	}
	if handled {
		t.Fatal("expected unknown command to be ignored")
	}
}
