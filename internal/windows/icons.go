package windows

import "quick_app_switcher/internal/win32"

type IconCache struct {
	icons map[string]win32.HICON
}

func NewIconCache() *IconCache {
	return &IconCache{icons: make(map[string]win32.HICON)}
}

func (c *IconCache) IconFor(info WindowInfo) win32.HICON {
	if c == nil {
		return 0
	}
	if icon := c.icons[info.ExecutablePath]; icon != 0 {
		return icon
	}
	icon := win32.LoadDefaultApplicationIcon()
	c.icons[info.ExecutablePath] = icon
	return icon
}
