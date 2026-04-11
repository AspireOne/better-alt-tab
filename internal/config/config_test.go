package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPath(t *testing.T) {
	t.Run("missing file returns defaults", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), configName)

		cfg, err := loadPath(path)
		if err != nil {
			t.Fatalf("loadPath returned error: %v", err)
		}
		if cfg != Default() {
			t.Fatalf("got %+v want %+v", cfg, Default())
		}
		// #nosec G304 -- The test controls the temp path.
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read created file: %v", err)
		}
		if !strings.Contains(string(data), "show_thumbnails = true") {
			t.Fatalf("got file contents %q", data)
		}
		if !strings.Contains(string(data), "launch_on_startup = false") {
			t.Fatalf("got file contents %q", data)
		}
	})

	t.Run("empty file keeps defaults", func(t *testing.T) {
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
	})

	t.Run("explicit value overrides defaults", func(t *testing.T) {
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
	})

	t.Run("invalid toml fails", func(t *testing.T) {
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
	})

	t.Run("unknown keys fail", func(t *testing.T) {
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
	})

	t.Run("save path writes round trip config", func(t *testing.T) {
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
	})
}
