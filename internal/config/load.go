package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, configSubdir, appDirName, configName), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	return loadPath(path)
}

func loadPath(path string) (Config, error) {
	cfg := Default()

	// #nosec G304 -- The config path comes from the app's fixed location or test-controlled temp files.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := savePath(path, cfg); err != nil {
				return cfg, fmt.Errorf("create default config %s: %w", path, err)
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	meta, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return Default(), fmt.Errorf("decode config %s: %w", path, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, 0, len(undecoded))
		for _, key := range undecoded {
			keys = append(keys, key.String())
		}
		return Default(), fmt.Errorf("decode config %s: unknown keys: %s", path, strings.Join(keys, ", "))
	}
	if err := savePath(path, cfg); err != nil {
		return cfg, fmt.Errorf("normalize config %s: %w", path, err)
	}
	return cfg, nil
}
