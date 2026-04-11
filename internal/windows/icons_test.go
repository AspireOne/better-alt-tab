package windows

import (
	"reflect"
	"testing"

	"better_alt_tab/internal/win32"
)

type fakeIconResolver struct {
	windowIcons map[win32.HWND]win32.HICON
	classIcons  map[win32.HWND]win32.HICON
	shellIcons  map[string]win32.HICON
	fallback    win32.HICON
	destroyed   []win32.HICON
	windowCalls int
	classCalls  int
	shellCalls  int
}

func (r *fakeIconResolver) WindowIcon(hwnd win32.HWND) win32.HICON {
	r.windowCalls++
	return r.windowIcons[hwnd]
}

func (r *fakeIconResolver) ClassIcon(hwnd win32.HWND) win32.HICON {
	r.classCalls++
	return r.classIcons[hwnd]
}

func (r *fakeIconResolver) ShellIcon(path string) (win32.HICON, bool) {
	r.shellCalls++
	icon := r.shellIcons[path]
	return icon, icon != 0
}

func (r *fakeIconResolver) DefaultIcon() win32.HICON {
	return r.fallback
}

func (r *fakeIconResolver) DestroyIcon(icon win32.HICON) {
	r.destroyed = append(r.destroyed, icon)
}

func TestIconForReturnsFallbackBeforeWarm(t *testing.T) {
	cache := newTestIconCache()

	got := cache.IconFor(WindowInfo{ID: 1, ExecutablePath: `C:\Apps\app.exe`})
	if got != 100 {
		t.Fatalf("IconFor returned %#x, want fallback %#x", got, win32.HICON(100))
	}
}

func TestWarmStoresWindowIcon(t *testing.T) {
	resolver := newFakeIconResolver()
	resolver.windowIcons[win32.HWND(1)] = 200
	cache := newIconCacheWithResolver(resolver)

	if !cache.Warm([]WindowInfo{{ID: 1, ExecutablePath: `C:\Apps\app.exe`}}) {
		t.Fatal("Warm returned false, want true")
	}
	if got := cache.IconFor(WindowInfo{ID: 1, ExecutablePath: `C:\Apps\app.exe`}); got != 200 {
		t.Fatalf("IconFor returned %#x, want window icon %#x", got, win32.HICON(200))
	}
}

func TestWarmFallsBackToClassIcon(t *testing.T) {
	resolver := newFakeIconResolver()
	resolver.classIcons[win32.HWND(1)] = 300
	cache := newIconCacheWithResolver(resolver)

	if !cache.Warm([]WindowInfo{{ID: 1, ExecutablePath: `C:\Apps\app.exe`}}) {
		t.Fatal("Warm returned false, want true")
	}
	if got := cache.IconFor(WindowInfo{ID: 1, ExecutablePath: `C:\Apps\app.exe`}); got != 300 {
		t.Fatalf("IconFor returned %#x, want class icon %#x", got, win32.HICON(300))
	}
}

func TestWarmStoresShellIconByCleanPath(t *testing.T) {
	resolver := newFakeIconResolver()
	resolver.shellIcons[`C:\Apps\app.exe`] = 400
	cache := newIconCacheWithResolver(resolver)

	if !cache.Warm([]WindowInfo{{ID: 1, ExecutablePath: `C:\Apps\.\app.exe`}}) {
		t.Fatal("Warm returned false, want true")
	}
	if got := cache.IconFor(WindowInfo{ID: 2, ExecutablePath: `C:\Apps\app.exe`}); got != 400 {
		t.Fatalf("IconFor returned %#x, want shell icon %#x", got, win32.HICON(400))
	}
}

func TestWarmReturnsFalseWhenAlreadyCached(t *testing.T) {
	cache := newTestIconCache()
	if !cache.storeWindow(1, 200, false) {
		t.Fatal("storeWindow returned false, want true")
	}

	if cache.Warm([]WindowInfo{{ID: 1, ExecutablePath: `C:\Apps\app.exe`}}) {
		t.Fatal("Warm returned true, want false")
	}
}

func TestWarmSkipsProbingWhenPathIconAlreadyCached(t *testing.T) {
	resolver := newFakeIconResolver()
	cache := newIconCacheWithResolver(resolver)
	if !cache.storePath(`C:\Apps\app.exe`, 400, true) {
		t.Fatal("storePath returned false, want true")
	}

	if cache.Warm([]WindowInfo{{ID: 1, ExecutablePath: `C:\Apps\app.exe`}}) {
		t.Fatal("Warm returned true, want false")
	}
	if resolver.windowCalls != 0 || resolver.classCalls != 0 || resolver.shellCalls != 0 {
		t.Fatalf("resolver calls = window:%d class:%d shell:%d, want all zero", resolver.windowCalls, resolver.classCalls, resolver.shellCalls)
	}
}

func TestDuplicatePathStoreDestroysOwnedLoser(t *testing.T) {
	resolver := newFakeIconResolver()
	cache := newIconCacheWithResolver(resolver)

	if !cache.storePath(`C:\Apps\app.exe`, 400, true) {
		t.Fatal("initial storePath returned false, want true")
	}
	if cache.storePath(`C:\Apps\app.exe`, 401, true) {
		t.Fatal("duplicate storePath returned true, want false")
	}
	if !reflect.DeepEqual(resolver.destroyed, []win32.HICON{401}) {
		t.Fatalf("destroyed icons = %v, want %v", resolver.destroyed, []win32.HICON{401})
	}
}

func TestDuplicateWindowStoreDoesNotDestroyBorrowedLoser(t *testing.T) {
	resolver := newFakeIconResolver()
	cache := newIconCacheWithResolver(resolver)

	if !cache.storeWindow(1, 200, false) {
		t.Fatal("initial storeWindow returned false, want true")
	}
	if cache.storeWindow(1, 201, false) {
		t.Fatal("duplicate storeWindow returned true, want false")
	}
	if len(resolver.destroyed) != 0 {
		t.Fatalf("destroyed icons = %v, want none", resolver.destroyed)
	}
}

func TestCloseDestroysOnlyOwnedIcons(t *testing.T) {
	resolver := newFakeIconResolver()
	cache := newIconCacheWithResolver(resolver)
	cache.storeWindow(1, 200, false)
	cache.storePath(`C:\Apps\app.exe`, 400, true)

	cache.Close()

	if !reflect.DeepEqual(resolver.destroyed, []win32.HICON{400}) {
		t.Fatalf("destroyed icons = %v, want %v", resolver.destroyed, []win32.HICON{400})
	}
	if got := cache.IconFor(WindowInfo{ID: 1, ExecutablePath: `C:\Apps\app.exe`}); got != 100 {
		t.Fatalf("IconFor returned %#x after Close, want fallback %#x", got, win32.HICON(100))
	}
}

func newTestIconCache() *IconCache {
	return newIconCacheWithResolver(newFakeIconResolver())
}

func newFakeIconResolver() *fakeIconResolver {
	return &fakeIconResolver{
		windowIcons: make(map[win32.HWND]win32.HICON),
		classIcons:  make(map[win32.HWND]win32.HICON),
		shellIcons:  make(map[string]win32.HICON),
		fallback:    100,
	}
}
