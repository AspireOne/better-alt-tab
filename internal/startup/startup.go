package startup

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	valueName  = "QuickAppSwitcher"
)

func Sync(enabled bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	command := commandLine(exePath)

	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open run key: %w", err)
	}

	if enabled {
		if err := key.SetStringValue(valueName, command); err != nil {
			_ = key.Close()
			return fmt.Errorf("set run value: %w", err)
		}
		if err := key.Close(); err != nil {
			return fmt.Errorf("close run key: %w", err)
		}
		return nil
	}

	if err := key.DeleteValue(valueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		_ = key.Close()
		return fmt.Errorf("delete run value: %w", err)
	}
	if err := key.Close(); err != nil {
		return fmt.Errorf("close run key: %w", err)
	}
	return nil
}

func commandLine(exePath string) string {
	return `"` + exePath + `"`
}
