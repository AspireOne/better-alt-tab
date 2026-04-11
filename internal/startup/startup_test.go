package startup

import "testing"

func TestCommandLine(t *testing.T) {
	t.Run("quotes path with spaces", func(t *testing.T) {
		got := commandLine(`C:\Program Files\Quick App Switcher\quick-app-switcher.exe`)
		want := `"C:\Program Files\Quick App Switcher\quick-app-switcher.exe"`
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("quotes path without spaces", func(t *testing.T) {
		got := commandLine(`C:\tools\quick-app-switcher.exe`)
		want := `"C:\tools\quick-app-switcher.exe"`
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}
