package ui

import "quick_app_switcher/internal/win32"

const CommandExit = 1001

func RegisterTaskbarCreated() uint32 {
	return win32.RegisterWindowMessage("TaskbarCreated")
}
