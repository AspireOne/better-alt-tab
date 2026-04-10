package ui

import "quick_app_switcher/internal/win32"

type OverlayMetrics struct {
	Width      int32
	Height     int32
	IconSize   int32
	Padding    int32
	ItemStride int32
}

func ComputeMetrics(count int) OverlayMetrics {
	metrics := OverlayMetrics{
		Height:     96,
		IconSize:   32,
		Padding:    16,
		ItemStride: 48,
	}
	if count < 1 {
		count = 1
	}
	metrics.Width = metrics.Padding * 2
	for i := 0; i < count; i++ {
		if metrics.Width > 0x7fffffff-metrics.ItemStride {
			metrics.Width = 0x7fffffff
			return metrics
		}
		metrics.Width += metrics.ItemStride
	}
	return metrics
}

func CenterRectOnWindow(anchor win32.HWND, metrics OverlayMetrics) win32.RECT {
	monitor := win32.MonitorFromWindow(anchor)
	bounds := win32.GetMonitorRect(monitor)
	left := bounds.Left + ((bounds.Right-bounds.Left)-metrics.Width)/2
	top := bounds.Top + ((bounds.Bottom-bounds.Top)-metrics.Height)/3
	return win32.RECT{
		Left:   left,
		Top:    top,
		Right:  left + metrics.Width,
		Bottom: top + metrics.Height,
	}
}
