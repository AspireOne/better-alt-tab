package ui

import (
	"testing"

	"better_alt_tab/internal/theme"
)

func TestFitMetricsToWidthCapsOverlayWidth(t *testing.T) {
	metrics := ComputeMetrics(theme.Default().Layout, true, 8)
	fitted := FitMetricsToWidth(metrics, 8, 900)
	if fitted.Width > 900 {
		t.Fatalf("overlay width exceeded bounds: got %d want <= 900", fitted.Width)
	}
	if fitted.ThumbnailWidth <= 0 {
		t.Fatalf("thumbnail width must stay positive, got %d", fitted.ThumbnailWidth)
	}
}

func TestFitMetricsToWidthPreservesDefaultWhenSpaceAllows(t *testing.T) {
	metrics := ComputeMetrics(theme.Default().Layout, true, 3)
	fitted := FitMetricsToWidth(metrics, 3, 2000)
	if fitted.ThumbnailWidth != metrics.ThumbnailWidth {
		t.Fatalf("unexpected thumbnail shrink: got %d want %d", fitted.ThumbnailWidth, metrics.ThumbnailWidth)
	}
	if fitted.Width != metrics.Width {
		t.Fatalf("unexpected width change: got %d want %d", fitted.Width, metrics.Width)
	}
}

func TestFitMetricsToWidthIncludesLabelSpace(t *testing.T) {
	metrics := ComputeMetrics(theme.Default().Layout, true, 2)
	want := metrics.Padding*2 + metrics.ThumbnailHeight + metrics.LabelGap + metrics.LabelHeight
	if metrics.Height != want {
		t.Fatalf("height = %d, want %d", metrics.Height, want)
	}
}

func TestFitMetricsToWidthCapsOverlayWidthForExtremeCounts(t *testing.T) {
	metrics := ComputeMetrics(theme.Default().Layout, true, 40)
	fitted := FitMetricsToWidth(metrics, 40, 900)
	if fitted.Width > 900 {
		t.Fatalf("overlay width exceeded bounds: got %d want <= 900", fitted.Width)
	}
}

func TestComputeMetricsWithoutLabelsIsCompact(t *testing.T) {
	metrics := ComputeMetrics(theme.Default().Layout, false, 2)
	if metrics.LabelGap != 0 {
		t.Fatalf("expected no label gap, got %d", metrics.LabelGap)
	}
	if metrics.LabelHeight != 0 {
		t.Fatalf("expected no label height, got %d", metrics.LabelHeight)
	}
	want := metrics.Padding*2 + metrics.ThumbnailHeight
	if metrics.Height != want {
		t.Fatalf("height = %d, want %d", metrics.Height, want)
	}
}
