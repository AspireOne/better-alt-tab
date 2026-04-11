package ui

import "testing"

func TestFitMetricsToWidthCapsOverlayWidth(t *testing.T) {
	metrics := ComputeMetrics(8)
	fitted := FitMetricsToWidth(metrics, 8, 900)
	if fitted.Width > 900 {
		t.Fatalf("overlay width exceeded bounds: got %d want <= 900", fitted.Width)
	}
	if fitted.ThumbnailWidth <= 0 {
		t.Fatalf("thumbnail width must stay positive, got %d", fitted.ThumbnailWidth)
	}
}

func TestFitMetricsToWidthPreservesDefaultWhenSpaceAllows(t *testing.T) {
	metrics := ComputeMetrics(3)
	fitted := FitMetricsToWidth(metrics, 3, 2000)
	if fitted.ThumbnailWidth != metrics.ThumbnailWidth {
		t.Fatalf("unexpected thumbnail shrink: got %d want %d", fitted.ThumbnailWidth, metrics.ThumbnailWidth)
	}
	if fitted.Width != metrics.Width {
		t.Fatalf("unexpected width change: got %d want %d", fitted.Width, metrics.Width)
	}
}

func TestFitMetricsToWidthCapsOverlayWidthForExtremeCounts(t *testing.T) {
	metrics := ComputeMetrics(40)
	fitted := FitMetricsToWidth(metrics, 40, 900)
	if fitted.Width > 900 {
		t.Fatalf("overlay width exceeded bounds: got %d want <= 900", fitted.Width)
	}
}
