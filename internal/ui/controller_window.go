package ui

import "better_alt_tab/internal/win32"

const (
	CommandExit           = 1001
	CommandOpenSettings   = 1002
	CommandOpenConfigFile = 1003
)

func RegisterTaskbarCreated() uint32 {
	return win32.RegisterWindowMessage("TaskbarCreated")
}
