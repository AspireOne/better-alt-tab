package ui

import (
	"better_alt_tab/internal/theme"
	"better_alt_tab/internal/win32"
)

type OverlayMetrics struct {
	Width           int32
	Height          int32
	ThumbnailWidth  int32
	ThumbnailHeight int32
	IconSize        int32
	LabelHeight     int32
	LabelGap        int32
	Padding         int32
	Gap             int32
	SelectionInset  int32
}

func ComputeMetrics(layout theme.Layout, showLabels bool, count int) OverlayMetrics {
	metrics := OverlayMetrics{
		ThumbnailWidth:  layout.ThumbnailWidth,
		ThumbnailHeight: layout.ThumbnailHeight,
		IconSize:        layout.IconSize,
		Padding:         layout.Padding,
		Gap:             layout.Gap,
		SelectionInset:  layout.SelectionInset,
	}
	if showLabels {
		metrics.LabelHeight = layout.LabelHeight
		metrics.LabelGap = layout.LabelGap
	}
	return FitMetricsToWidth(metrics, count, 0x7fffffff)
}

func ComputeMetricsForAnchor(anchor win32.HWND, layout theme.Layout, showLabels bool, count int) OverlayMetrics {
	monitor := win32.MonitorFromWindow(anchor)
	bounds := win32.GetMonitorRect(monitor)
	return FitMetricsToWidth(ComputeMetrics(layout, showLabels, count), count, bounds.Right-bounds.Left)
}

func FitMetricsToWidth(metrics OverlayMetrics, count int, maxWidth int32) OverlayMetrics {
	count32 := metricCount(count)
	if maxWidth > 0 {
		gap := metrics.Gap
		thumbWidth := metrics.ThumbnailWidth
		const minGap int32 = 6
		const minThumbWidth int32 = 48

		maxThumbWidth := (maxWidth - metrics.Padding*2 - gap*(count32-1)) / count32
		if maxThumbWidth < thumbWidth {
			thumbWidth = maxThumbWidth
		}
		if thumbWidth < minThumbWidth {
			gap = minGap
			maxThumbWidth = (maxWidth - metrics.Padding*2 - gap*(count32-1)) / count32
			if maxThumbWidth < thumbWidth {
				thumbWidth = maxThumbWidth
			}
		}
		if thumbWidth < minThumbWidth {
			thumbWidth = minThumbWidth
		}
		finalWidth := metrics.Padding*2 + thumbWidth*count32 + gap*(count32-1)
		if finalWidth > maxWidth {
			thumbWidth = (maxWidth - metrics.Padding*2 - gap*(count32-1)) / count32
			if thumbWidth < 1 {
				thumbWidth = 1
			}
		}
		if thumbWidth < metrics.ThumbnailWidth {
			metrics.ThumbnailWidth = thumbWidth
			metrics.ThumbnailHeight = (thumbWidth * metrics.ThumbnailHeight) / 180
			if metrics.ThumbnailHeight < 40 {
				metrics.ThumbnailHeight = 40
			}
			metrics.Gap = gap
			iconSize := thumbWidth / 6
			if iconSize < 16 {
				iconSize = 16
			}
			if iconSize < metrics.IconSize {
				metrics.IconSize = iconSize
			}
		}
	}
	metrics.Width = metrics.Padding*2 + metrics.ThumbnailWidth*count32 + metrics.Gap*(count32-1)
	metrics.Height = metrics.Padding*2 + metrics.ThumbnailHeight + metrics.LabelGap + metrics.LabelHeight
	return metrics
}

func metricCount(count int) int32 {
	const maxInt32 = 1<<31 - 1
	if count < 1 {
		return 1
	}
	if count > maxInt32 {
		return maxInt32
	}
	return int32(count) //nolint:gosec // range checked above
}

func metricIndex(index int) int32 {
	const maxInt32 = 1<<31 - 1
	if index < 0 {
		return 0
	}
	if index > maxInt32 {
		return maxInt32
	}
	return int32(index) //nolint:gosec // range checked above
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
