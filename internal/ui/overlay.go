package ui

import (
	"quick_app_switcher/internal/win32"
	"quick_app_switcher/internal/windows"
)

type Overlay struct {
	hwnd win32.HWND
	data OverlayData
}

type OverlayData struct {
	Items    []windows.WindowInfo
	Selected int
}

func NewOverlay(hwnd win32.HWND) *Overlay {
	return &Overlay{hwnd: hwnd}
}

func (o *Overlay) Update(anchor win32.HWND, items []windows.WindowInfo, selected int) {
	o.data.Items = append(o.data.Items[:0], items...)
	o.data.Selected = selected
	metrics := ComputeMetrics(len(items))
	rect := CenterRectOnWindow(anchor, metrics)
	win32.PositionWindowNoActivate(o.hwnd, rect.Left, rect.Top, metrics.Width, metrics.Height, true)
	win32.InvalidateRect(o.hwnd)
}

func (o *Overlay) Hide() {
	win32.ShowWindow(o.hwnd, win32.SW_HIDE)
}

func (o *Overlay) Paint(hwnd win32.HWND, icons *windows.IconCache) {
	ps, hdc := win32.BeginPaint(hwnd)
	defer win32.EndPaint(hwnd, &ps)

	rect := ps.Paint
	bg := win32.CreateSolidBrush(0x00202020)
	defer win32.DeleteObject(uintptr(bg))
	win32.FillRect(hdc, &rect, bg)

	metrics := ComputeMetrics(len(o.data.Items))
	startX := metrics.Padding
	left := startX
	for i, item := range o.data.Items {
		itemRect := win32.RECT{
			Left:   left,
			Top:    metrics.Padding,
			Right:  left + metrics.IconSize,
			Bottom: metrics.Padding + metrics.IconSize,
		}
		fill := uintptr(0x00444444)
		if i == o.data.Selected {
			fill = 0x00c06020
		}
		brush := win32.CreateSolidBrush(fill)
		win32.FillRect(hdc, &itemRect, brush)
		win32.DeleteObject(uintptr(brush))
		icon := icons.IconFor(item)
		win32.DrawIconInRect(hdc, itemRect, icon)
		left += metrics.ItemStride
	}
}
