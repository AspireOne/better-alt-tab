package windows

import (
	"errors"
	"sync"

	"quick_app_switcher/internal/win32"
)

type Thumbnail struct {
	Bitmap win32.HBITMAP
	Width  int32
	Height int32
	State  ThumbnailState
}

type ThumbnailState uint8

const (
	ThumbnailStatePreview ThumbnailState = iota + 1
	ThumbnailStateFallback
)

type ThumbnailCache struct {
	mu    sync.RWMutex
	items map[WindowID]Thumbnail
}

func NewThumbnailCache() *ThumbnailCache {
	return &ThumbnailCache{items: make(map[WindowID]Thumbnail)}
}

func (c *ThumbnailCache) ThumbnailFor(id WindowID) (Thumbnail, bool) {
	if c == nil {
		return Thumbnail{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	thumb, ok := c.items[id]
	return thumb, ok
}

func (c *ThumbnailCache) WithThumbnail(id WindowID, fn func(Thumbnail) bool) bool {
	if c == nil || fn == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	thumb, ok := c.items[id]
	if !ok {
		return false
	}
	return fn(thumb)
}

func (c *ThumbnailCache) Warm(items []WindowInfo, targetWidth, targetHeight int32) {
	if c == nil || targetWidth <= 0 || targetHeight <= 0 {
		return
	}
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		if item.Minimized {
			c.mu.Lock()
			c.setFallbackLocked(item.ID, targetWidth, targetHeight)
			c.mu.Unlock()
			continue
		}
		thumb, err := captureThumbnail(item.ID.HWND(), targetWidth, targetHeight)
		if err != nil {
			continue
		}
		c.mu.Lock()
		if existing, ok := c.items[item.ID]; ok && existing.Bitmap != 0 {
			win32.DeleteObject(uintptr(existing.Bitmap))
		}
		c.items[item.ID] = thumb
		c.mu.Unlock()
	}
}

func (c *ThumbnailCache) setFallbackLocked(id WindowID, width, height int32) {
	existing, ok := c.items[id]
	if ok && existing.State == ThumbnailStatePreview && existing.Bitmap != 0 {
		return
	}
	c.items[id] = Thumbnail{
		Width:  width,
		Height: height,
		State:  ThumbnailStateFallback,
	}
}

func (c *ThumbnailCache) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, thumb := range c.items {
		if thumb.Bitmap != 0 {
			win32.DeleteObject(uintptr(thumb.Bitmap))
		}
		delete(c.items, id)
	}
}

func captureThumbnail(hwnd win32.HWND, targetWidth, targetHeight int32) (Thumbnail, error) {
	if hwnd == 0 || targetWidth <= 0 || targetHeight <= 0 {
		return Thumbnail{}, errors.New("invalid thumbnail request")
	}
	rect, ok := win32.GetWindowRect(hwnd)
	if !ok {
		return Thumbnail{}, errors.New("read window rect")
	}
	srcWidth := rect.Right - rect.Left
	srcHeight := rect.Bottom - rect.Top
	if srcWidth <= 0 || srcHeight <= 0 {
		return Thumbnail{}, errors.New("window has no size")
	}

	screenDC := win32.GetDC(0)
	if screenDC == 0 {
		return Thumbnail{}, errors.New("get screen dc")
	}
	defer win32.ReleaseDC(0, screenDC)

	srcDC := win32.CreateCompatibleDC(screenDC)
	if srcDC == 0 {
		return Thumbnail{}, errors.New("create source dc")
	}
	defer win32.DeleteDC(srcDC)

	srcBitmap := win32.CreateCompatibleBitmap(screenDC, srcWidth, srcHeight)
	if srcBitmap == 0 {
		return Thumbnail{}, errors.New("create source bitmap")
	}
	defer win32.DeleteObject(uintptr(srcBitmap))
	oldSrc := win32.SelectObject(srcDC, win32.HGDIOBJ(srcBitmap))
	defer win32.SelectObject(srcDC, oldSrc)

	rendered := win32.PrintWindow(hwnd, srcDC, win32.PW_RENDERFULLCONTENT)
	if !rendered && !win32.IsIconic(hwnd) {
		windowDC := win32.GetWindowDC(hwnd)
		if windowDC != 0 {
			rendered = win32.BitBlt(srcDC, 0, 0, srcWidth, srcHeight, windowDC, 0, 0, win32.SRCCOPY|win32.CAPTUREBLT)
			win32.ReleaseDC(hwnd, windowDC)
		}
	}
	if !rendered {
		return Thumbnail{}, errors.New("capture window contents")
	}

	dstDC := win32.CreateCompatibleDC(screenDC)
	if dstDC == 0 {
		return Thumbnail{}, errors.New("create destination dc")
	}
	defer win32.DeleteDC(dstDC)

	dstBitmap := win32.CreateCompatibleBitmap(screenDC, targetWidth, targetHeight)
	if dstBitmap == 0 {
		return Thumbnail{}, errors.New("create destination bitmap")
	}
	keepDstBitmap := false
	defer func() {
		if !keepDstBitmap {
			win32.DeleteObject(uintptr(dstBitmap))
		}
	}()
	oldDst := win32.SelectObject(dstDC, win32.HGDIOBJ(dstBitmap))
	defer win32.SelectObject(dstDC, oldDst)

	win32.SetStretchBltMode(dstDC, win32.HALFTONE)
	if !win32.StretchBlt(dstDC, 0, 0, targetWidth, targetHeight, srcDC, 0, 0, srcWidth, srcHeight, win32.SRCCOPY) {
		return Thumbnail{}, errors.New("scale thumbnail")
	}
	keepDstBitmap = true

	return Thumbnail{
		Bitmap: dstBitmap,
		Width:  targetWidth,
		Height: targetHeight,
		State:  ThumbnailStatePreview,
	}, nil
}
