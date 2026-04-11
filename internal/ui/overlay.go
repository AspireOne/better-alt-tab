package ui

import (
	"better_alt_tab/internal/win32"
	"better_alt_tab/internal/windows"
)

type Overlay struct {
	hwnd    win32.HWND
	data    OverlayData
	metrics OverlayMetrics
}

type OverlayData struct {
	Items    []windows.WindowInfo
	Selected int
}

func NewOverlay(hwnd win32.HWND) *Overlay {
	return &Overlay{hwnd: hwnd}
}

func (o *Overlay) Update(anchor win32.HWND, items []windows.WindowInfo, selected int) {
	metrics := ComputeMetricsForAnchor(anchor, len(items))
	o.UpdateWithMetrics(anchor, items, selected, metrics)
}

func (o *Overlay) UpdateWithMetrics(anchor win32.HWND, items []windows.WindowInfo, selected int, metrics OverlayMetrics) {
	o.data.Items = append(o.data.Items[:0], items...)
	o.data.Selected = selected
	o.metrics = metrics
	rect := CenterRectOnWindow(anchor, metrics)
	win32.PositionWindowNoActivate(o.hwnd, rect.Left, rect.Top, metrics.Width, metrics.Height, true)
	win32.InvalidateRect(o.hwnd)
}

func (o *Overlay) UpdateSelection(selected int) {
	if selected == o.data.Selected {
		return
	}
	previous := o.data.Selected
	o.data.Selected = selected
	if o.metrics.Width == 0 || o.metrics.Height == 0 {
		win32.InvalidateRect(o.hwnd)
		return
	}
	o.invalidateItem(previous)
	o.invalidateItem(selected)
}

func (o *Overlay) RefreshThumbnails() {
	o.Refresh()
}

func (o *Overlay) Refresh() {
	win32.InvalidateRect(o.hwnd)
}

func (o *Overlay) Hide() {
	win32.ShowWindow(o.hwnd, win32.SW_HIDE)
}

func (o *Overlay) Paint(hwnd win32.HWND, icons *windows.IconCache, thumbnails *windows.ThumbnailCache) {
	ps, hdc := win32.BeginPaint(hwnd)
	defer win32.EndPaint(hwnd, &ps)

	rect := ps.Paint
	bg := win32.CreateSolidBrush(0x00202020)
	defer win32.DeleteObject(uintptr(bg))
	win32.FillRect(hdc, &rect, bg)

	metrics := o.metrics
	if metrics.Width == 0 || metrics.Height == 0 {
		metrics = ComputeMetrics(len(o.data.Items))
	}
	left := metrics.Padding
	for i, item := range o.data.Items {
		selectionRect := win32.RECT{
			Left:   left - metrics.SelectionInset,
			Top:    metrics.Padding - metrics.SelectionInset,
			Right:  left + metrics.ThumbnailWidth + metrics.SelectionInset,
			Bottom: metrics.Padding + metrics.ThumbnailHeight + metrics.LabelGap + metrics.LabelHeight + metrics.SelectionInset,
		}
		if !rectsIntersect(rect, selectionRect) {
			left += metrics.ThumbnailWidth + metrics.Gap
			continue
		}
		fill := uintptr(0x00333333)
		if i == o.data.Selected {
			fill = 0x00c06020
		}
		brush := win32.CreateSolidBrush(fill)
		win32.FillRect(hdc, &selectionRect, brush)
		win32.DeleteObject(uintptr(brush))

		thumbRect := win32.RECT{
			Left:   left,
			Top:    metrics.Padding,
			Right:  left + metrics.ThumbnailWidth,
			Bottom: metrics.Padding + metrics.ThumbnailHeight,
		}
		drawFallbackPreview := true
		if thumbnails != nil {
			drawn := thumbnails.WithThumbnail(item.ID, func(thumb windows.Thumbnail) bool {
				if thumb.State != windows.ThumbnailStatePreview || thumb.Bitmap == 0 {
					return false
				}
				return win32.DrawBitmapInRect(hdc, thumbRect, thumb.Bitmap, thumb.Width, thumb.Height)
			})
			if drawn {
				drawFallbackPreview = false
			}
		}
		if drawFallbackPreview {
			fallback := win32.CreateSolidBrush(0x00444444)
			win32.FillRect(hdc, &thumbRect, fallback)
			win32.DeleteObject(uintptr(fallback))

			fallbackIconSize := metrics.IconSize * 3
			if fallbackIconSize > metrics.ThumbnailWidth-24 {
				fallbackIconSize = metrics.ThumbnailWidth - 24
			}
			if fallbackIconSize > metrics.ThumbnailHeight-24 {
				fallbackIconSize = metrics.ThumbnailHeight - 24
			}
			if fallbackIconSize < metrics.IconSize {
				fallbackIconSize = metrics.IconSize
			}
			fallbackIconRect := win32.RECT{
				Left:   thumbRect.Left + (metrics.ThumbnailWidth-fallbackIconSize)/2,
				Top:    thumbRect.Top + (metrics.ThumbnailHeight-fallbackIconSize)/2,
				Right:  thumbRect.Left + (metrics.ThumbnailWidth+fallbackIconSize)/2,
				Bottom: thumbRect.Top + (metrics.ThumbnailHeight+fallbackIconSize)/2,
			}
			win32.DrawIconInRect(hdc, fallbackIconRect, icons.IconFor(item))
		}

		badgeRect := win32.RECT{
			Left:   thumbRect.Left + 8,
			Top:    thumbRect.Top + 8,
			Right:  thumbRect.Left + 8 + metrics.IconSize + 8,
			Bottom: thumbRect.Top + 8 + metrics.IconSize + 8,
		}
		badge := win32.CreateSolidBrush(0x00181818)
		win32.FillRect(hdc, &badgeRect, badge)
		win32.DeleteObject(uintptr(badge))
		iconRect := win32.RECT{
			Left:   badgeRect.Left + 4,
			Top:    badgeRect.Top + 4,
			Right:  badgeRect.Right - 4,
			Bottom: badgeRect.Bottom - 4,
		}
		win32.DrawIconInRect(hdc, iconRect, icons.IconFor(item))

		labelRect := win32.RECT{
			Left:   thumbRect.Left,
			Top:    thumbRect.Bottom + metrics.LabelGap,
			Right:  thumbRect.Right,
			Bottom: thumbRect.Bottom + metrics.LabelGap + metrics.LabelHeight,
		}
		labelColor := uintptr(0x00d8d8d8)
		if i == o.data.Selected {
			labelColor = 0x00ffffff
		}
		win32.DrawLabel(hdc, labelRect, item.AppDisplayName(), labelColor)

		left += metrics.ThumbnailWidth + metrics.Gap
	}
}

func (o *Overlay) invalidateItem(index int) {
	if index < 0 || index >= len(o.data.Items) {
		return
	}
	win32.InvalidateRectArea(o.hwnd, itemSelectionRect(index, o.metrics))
}

func itemSelectionRect(index int, metrics OverlayMetrics) win32.RECT {
	left := metrics.Padding + metricIndex(index)*(metrics.ThumbnailWidth+metrics.Gap)
	return win32.RECT{
		Left:   left - metrics.SelectionInset,
		Top:    metrics.Padding - metrics.SelectionInset,
		Right:  left + metrics.ThumbnailWidth + metrics.SelectionInset,
		Bottom: metrics.Padding + metrics.ThumbnailHeight + metrics.LabelGap + metrics.LabelHeight + metrics.SelectionInset,
	}
}

func rectsIntersect(a, b win32.RECT) bool {
	return a.Left < b.Right && a.Right > b.Left && a.Top < b.Bottom && a.Bottom > b.Top
}
