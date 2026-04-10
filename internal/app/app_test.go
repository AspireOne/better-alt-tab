package app

import "testing"

func TestTrayNotificationCodeUsesLowWord(t *testing.T) {
	got := trayNotificationCode(0x1234007B)
	if got != wmContextMenu {
		t.Fatalf("got %#x want %#x", got, wmContextMenu)
	}
}
