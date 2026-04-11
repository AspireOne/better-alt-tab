package windows

import (
	"path/filepath"
	"sync"

	"better_alt_tab/internal/win32"
)

type IconCache struct {
	mu       sync.RWMutex
	warmMu   sync.Mutex
	byWindow map[WindowID]iconEntry
	byPath   map[string]iconEntry
	fallback win32.HICON
	resolver iconResolver
	closed   bool
}

type iconEntry struct {
	handle win32.HICON
	owned  bool
}

type iconResolver interface {
	WindowIcon(win32.HWND) win32.HICON
	ClassIcon(win32.HWND) win32.HICON
	ShellIcon(string) (win32.HICON, bool)
	DefaultIcon() win32.HICON
	DestroyIcon(win32.HICON)
}

type win32IconResolver struct{}

func NewIconCache() *IconCache {
	return newIconCacheWithResolver(win32IconResolver{})
}

func newIconCacheWithResolver(resolver iconResolver) *IconCache {
	return &IconCache{
		byWindow: make(map[WindowID]iconEntry),
		byPath:   make(map[string]iconEntry),
		fallback: resolver.DefaultIcon(),
		resolver: resolver,
	}
}

func (c *IconCache) IconFor(info WindowInfo) win32.HICON {
	if c == nil {
		return 0
	}
	key := pathKey(info.ExecutablePath)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, ok := c.byWindow[info.ID]; ok && entry.handle != 0 {
		return entry.handle
	}
	if key != "" {
		if entry, ok := c.byPath[key]; ok && entry.handle != 0 {
			return entry.handle
		}
	}
	return c.fallback
}

func (c *IconCache) Warm(items []WindowInfo) bool {
	if c == nil {
		return false
	}
	c.warmMu.Lock()
	defer c.warmMu.Unlock()

	if c.isClosed() {
		return false
	}

	changed := false
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		key := pathKey(item.ExecutablePath)
		if c.hasWindow(item.ID) || (key != "" && c.hasPath(key)) {
			continue
		}

		if icon := c.resolver.WindowIcon(item.ID.HWND()); icon != 0 {
			changed = c.storeWindow(item.ID, icon, false) || changed
			continue
		}
		if icon := c.resolver.ClassIcon(item.ID.HWND()); icon != 0 {
			changed = c.storeWindow(item.ID, icon, false) || changed
			continue
		}
		if key != "" {
			if icon, ok := c.resolver.ShellIcon(key); ok && icon != 0 {
				changed = c.storePath(key, icon, true) || changed
			}
		}
	}
	return changed
}

func (c *IconCache) Close() {
	if c == nil {
		return
	}
	c.warmMu.Lock()
	defer c.warmMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true
	for id, entry := range c.byWindow {
		if entry.owned {
			c.resolver.DestroyIcon(entry.handle)
		}
		delete(c.byWindow, id)
	}
	for key, entry := range c.byPath {
		if entry.owned {
			c.resolver.DestroyIcon(entry.handle)
		}
		delete(c.byPath, key)
	}
}

func (c *IconCache) hasWindow(id WindowID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.byWindow[id]; ok {
		return true
	}
	return false
}

func (c *IconCache) hasPath(key string) bool {
	if key == "" {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.byPath[key]
	return ok
}

func (c *IconCache) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

func (c *IconCache) storeWindow(id WindowID, icon win32.HICON, owned bool) bool {
	if id == 0 || icon == 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		c.destroyOwned(icon, owned)
		return false
	}
	if _, ok := c.byWindow[id]; ok {
		c.destroyOwned(icon, owned)
		return false
	}
	c.byWindow[id] = iconEntry{handle: icon, owned: owned}
	return true
}

func (c *IconCache) storePath(key string, icon win32.HICON, owned bool) bool {
	if key == "" || icon == 0 {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		c.destroyOwned(icon, owned)
		return false
	}
	if _, ok := c.byPath[key]; ok {
		c.destroyOwned(icon, owned)
		return false
	}
	c.byPath[key] = iconEntry{handle: icon, owned: owned}
	return true
}

func (c *IconCache) destroyOwned(icon win32.HICON, owned bool) {
	if owned && icon != 0 {
		c.resolver.DestroyIcon(icon)
	}
}

func pathKey(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func (win32IconResolver) WindowIcon(hwnd win32.HWND) win32.HICON {
	return win32.GetWindowIcon(hwnd)
}

func (win32IconResolver) ClassIcon(hwnd win32.HWND) win32.HICON {
	return win32.GetClassIcon(hwnd)
}

func (win32IconResolver) ShellIcon(path string) (win32.HICON, bool) {
	return win32.GetShellIcon(path)
}

func (win32IconResolver) DefaultIcon() win32.HICON {
	return win32.LoadDefaultApplicationIcon()
}

func (win32IconResolver) DestroyIcon(icon win32.HICON) {
	win32.DestroyIcon(icon)
}
