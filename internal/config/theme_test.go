package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"better_alt_tab/internal/theme"
)

func TestLoadThemePathMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "default.toml")

	got, err := loadThemePath(path)
	if err != nil {
		t.Fatalf("loadThemePath returned error: %v", err)
	}
	if got != theme.Default() {
		t.Fatalf("got %+v want %+v", got, theme.Default())
	}

	data := readConfigFile(t, path)
	if !strings.Contains(data, "overlay_background = \"#202020\"") {
		t.Fatalf("got theme contents %q", data)
	}
}

func TestLoadThemePathExplicitValuesOverrideDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sunset.toml")
	raw := "" +
		"version = 1\n" +
		"\n" +
		"[window]\n" +
		"opacity = 220\n" +
		"\n" +
		"[colors]\n" +
		"overlay_background = \"#112233\"\n" +
		"\n" +
		"[layout]\n" +
		"thumbnail_width = 200\n" +
		"\n" +
		"[features]\n" +
		"show_labels = false\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := loadThemePath(path)
	if err != nil {
		t.Fatalf("loadThemePath returned error: %v", err)
	}
	if got.Window.Opacity != 220 {
		t.Fatalf("got opacity %d want 220", got.Window.Opacity)
	}
	if got.Colors.OverlayBackground != 0x00332211 {
		t.Fatalf("got overlay background %#x", got.Colors.OverlayBackground)
	}
	if got.Layout.ThumbnailWidth != 200 {
		t.Fatalf("got thumbnail width %d want 200", got.Layout.ThumbnailWidth)
	}
	if got.Features.ShowLabels {
		t.Fatal("expected labels to be disabled")
	}
}

func TestLoadThemePathMissingFieldsKeepDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.toml")
	raw := "" +
		"version = 1\n" +
		"\n" +
		"[colors]\n" +
		"overlay_background = \"#112233\"\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := loadThemePath(path)
	if err != nil {
		t.Fatalf("loadThemePath returned error: %v", err)
	}
	if got.Layout != theme.Default().Layout {
		t.Fatalf("layout got %+v want %+v", got.Layout, theme.Default().Layout)
	}
	if got.Features != theme.Default().Features {
		t.Fatalf("features got %+v want %+v", got.Features, theme.Default().Features)
	}
}

func TestLoadThemePathInvalidLayoutFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.toml")
	raw := "" +
		"version = 1\n" +
		"\n" +
		"[layout]\n" +
		"thumbnail_width = 10\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := loadThemePath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "layout.thumbnail_width") {
		t.Fatalf("got error %q", err)
	}
	if got != theme.Default() {
		t.Fatalf("got %+v want %+v", got, theme.Default())
	}
}

func TestLoadThemePathUnknownKeysFail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.toml")
	if err := os.WriteFile(path, []byte("version = 1\nunknown = true\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := loadThemePath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown keys: unknown") {
		t.Fatalf("got error %q", err)
	}
	if got != theme.Default() {
		t.Fatalf("got %+v want %+v", got, theme.Default())
	}
}

func TestLoadThemePathInvalidColorFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.toml")
	raw := "" +
		"version = 1\n" +
		"\n" +
		"[colors]\n" +
		"overlay_background = \"blue\"\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := loadThemePath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "colors.overlay_background") {
		t.Fatalf("got error %q", err)
	}
	if got != theme.Default() {
		t.Fatalf("got %+v want %+v", got, theme.Default())
	}
}
