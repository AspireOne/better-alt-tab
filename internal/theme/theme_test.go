package theme

import "testing"

func TestNormalizeFillsMissingValues(t *testing.T) {
	got := (Theme{}).Normalize()
	if got != (Theme{}) {
		t.Fatalf("expected normalize to be a no-op, got %+v", got)
	}
}

func TestValidateRejectsInvalidLayout(t *testing.T) {
	cfg := Default()
	cfg.Layout.ThumbnailWidth = 10
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNormalizeKeepsExplicitFeatureFlags(t *testing.T) {
	cfg := Default()
	cfg.Features.ShowIconBadge = false
	cfg.Features.ShowLabels = false

	normalized := cfg.Normalize()
	if normalized.Features.ShowIconBadge {
		t.Fatal("expected show icon badge to stay disabled")
	}
	if normalized.Features.ShowLabels {
		t.Fatal("expected show labels to stay disabled")
	}
}
