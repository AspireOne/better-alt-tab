package ui

import "better_alt_tab/internal/win32"

type Tray struct {
	messageID uint32
	icon      win32.HICON
	tooltip   string
}

func NewTray(messageID uint32, tooltip string) *Tray {
	return &Tray{
		messageID: messageID,
		icon:      win32.LoadDefaultApplicationIcon(),
		tooltip:   tooltip,
	}
}

func (t *Tray) Add(hwnd win32.HWND) error {
	return win32.AddTrayIcon(hwnd, t.messageID, t.icon, t.tooltip)
}

func (t *Tray) Delete(hwnd win32.HWND) error {
	return win32.DeleteTrayIcon(hwnd)
}

func (t *Tray) ShowMenu(hwnd win32.HWND) {
	win32.ShowTrayMenu(hwnd, CommandOpenSettings, CommandOpenConfigFile, CommandReloadTheme, CommandExit)
}
