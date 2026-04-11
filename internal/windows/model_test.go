package windows

import "testing"

func TestWindowInfoAppDisplayNameUsesExecutableBase(t *testing.T) {
	info := WindowInfo{
		ExecutablePath: `C:\Program Files\App Folder\Example App.exe`,
		Title:          "Document.txt",
		ClassName:      "ExampleWindow",
	}

	if got := info.AppDisplayName(); got != "Example App" {
		t.Fatalf("display name = %q, want %q", got, "Example App")
	}
}

func TestWindowInfoAppDisplayNameFallsBackToTitleThenClass(t *testing.T) {
	withTitle := WindowInfo{Title: "Untitled - Editor", ClassName: "EditorWindow"}
	if got := withTitle.AppDisplayName(); got != "Untitled - Editor" {
		t.Fatalf("display name with title = %q, want %q", got, "Untitled - Editor")
	}

	withClass := WindowInfo{ClassName: "EditorWindow"}
	if got := withClass.AppDisplayName(); got != "EditorWindow" {
		t.Fatalf("display name with class = %q, want %q", got, "EditorWindow")
	}
}
