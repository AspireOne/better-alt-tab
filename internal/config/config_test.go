package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPathMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)

	cfg, err := loadPath(path)
	if err != nil {
		t.Fatalf("loadPath returned error: %v", err)
	}
	if cfg != Default() {
		t.Fatalf("got %+v want %+v", cfg, Default())
	}

	data := readConfigFile(t, path)
	if !strings.Contains(data, "show_thumbnails = true") {
		t.Fatalf("got file contents %q", data)
	}
	if !strings.Contains(data, "launch_on_startup = false") {
		t.Fatalf("got file contents %q", data)
	}
}

func TestLoadPathEmptyFileKeepsDefaultsAndNormalizes(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := loadPath(path)
	if err != nil {
		t.Fatalf("loadPath returned error: %v", err)
	}
	if cfg != Default() {
		t.Fatalf("got %+v want %+v", cfg, Default())
	}

	data := readConfigFile(t, path)
	if !strings.Contains(data, "show_thumbnails = true") {
		t.Fatalf("got file contents %q", data)
	}
	if !strings.Contains(data, "launch_on_startup = false") {
		t.Fatalf("got file contents %q", data)
	}
}

func TestLoadPathExplicitValuesOverrideDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	if err := os.WriteFile(path, []byte("show_thumbnails = false\nlaunch_on_startup = true\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := loadPath(path)
	if err != nil {
		t.Fatalf("loadPath returned error: %v", err)
	}
	if cfg.ShowThumbnails {
		t.Fatal("expected show_thumbnails to be false")
	}
	if !cfg.LaunchOnStartup {
		t.Fatal("expected launch_on_startup to be true")
	}
}

func TestLoadPathMissingFieldsAreAddedWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	if err := os.WriteFile(path, []byte("show_thumbnails = false\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := loadPath(path)
	if err != nil {
		t.Fatalf("loadPath returned error: %v", err)
	}
	want := Config{ShowThumbnails: false, LaunchOnStartup: false}
	if cfg != want {
		t.Fatalf("got %+v want %+v", cfg, want)
	}

	data := readConfigFile(t, path)
	if !strings.Contains(data, "show_thumbnails = false") {
		t.Fatalf("got file contents %q", data)
	}
	if !strings.Contains(data, "launch_on_startup = false") {
		t.Fatalf("got file contents %q", data)
	}
}

func TestLoadPathInvalidTOMLFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	if err := os.WriteFile(path, []byte("show_thumbnails =\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := loadPath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "decode config") {
		t.Fatalf("got error %q", err)
	}
	if cfg != Default() {
		t.Fatalf("got %+v want %+v", cfg, Default())
	}
}

func TestLoadPathUnknownKeysFail(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	if err := os.WriteFile(path, []byte("unknown = true\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := loadPath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown keys: unknown") {
		t.Fatalf("got error %q", err)
	}
	if cfg != Default() {
		t.Fatalf("got %+v want %+v", cfg, Default())
	}
}

func TestSavePathWritesRoundTripConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), configName)
	want := Config{ShowThumbnails: false, LaunchOnStartup: true}
	if err := savePath(path, want); err != nil {
		t.Fatalf("savePath returned error: %v", err)
	}

	got, err := loadPath(path)
	if err != nil {
		t.Fatalf("loadPath returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func readConfigFile(t *testing.T, path string) string {
	t.Helper()

	// #nosec G304 -- The test controls the temp path.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	return string(data)
}
