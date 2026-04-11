package ui

import "quick_app_switcher/internal/win32"

const (
	CommandExit           = 1001
	CommandOpenSettings   = 1002
	CommandOpenConfigFile = 1003
)

func RegisterTaskbarCreated() uint32 {
	return win32.RegisterWindowMessage("TaskbarCreated")
}
