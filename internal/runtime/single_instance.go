package runtime

import (
	"fmt"

	"quick_app_switcher/internal/win32"
)

const instanceMutexName = "Local\\QuickAppSwitcher.SingleInstance"

type SingleInstance struct {
	handle win32.Handle
}

func AcquireSingleInstance() (*SingleInstance, error) {
	handle, alreadyExists, err := win32.CreateNamedMutex(instanceMutexName)
	if err != nil {
		return nil, fmt.Errorf("create single-instance mutex: %w", err)
	}
	if alreadyExists {
		_ = win32.CloseHandle(handle)
		return nil, fmt.Errorf("quick app switcher is already running")
	}
	return &SingleInstance{handle: handle}, nil
}

func (s *SingleInstance) Release() error {
	if s == nil || s.handle == 0 {
		return nil
	}
	err := win32.CloseHandle(s.handle)
	s.handle = 0
	return err
}
