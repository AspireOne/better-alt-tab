package theme

import "fmt"

type Theme struct {
	Window   Window
	Colors   Colors
	Layout   Layout
	Features Features
}

type Window struct {
	Opacity byte
}

type Colors struct {
	OverlayBackground           uint32
	ItemBackground              uint32
	ItemSelectedBackground      uint32
	ThumbnailFallbackBackground uint32
	IconBadgeBackground         uint32
	Label                       uint32
	LabelSelected               uint32
}

type Layout struct {
	ThumbnailWidth  int32
	ThumbnailHeight int32
	IconSize        int32
	LabelHeight     int32
	LabelGap        int32
	Padding         int32
	Gap             int32
	SelectionInset  int32
}

type Features struct {
	ShowIconBadge bool
	ShowLabels    bool
}

func Default() Theme {
	return Theme{
		Window: Window{
			Opacity: 255,
		},
		Colors: Colors{
			OverlayBackground:           0x00202020,
			ItemBackground:              0x00333333,
			ItemSelectedBackground:      0x00c06020,
			ThumbnailFallbackBackground: 0x00444444,
			IconBadgeBackground:         0x00181818,
			Label:                       0x00d8d8d8,
			LabelSelected:               0x00ffffff,
		},
		Layout: Layout{
			ThumbnailWidth:  180,
			ThumbnailHeight: 110,
			IconSize:        20,
			LabelHeight:     18,
			LabelGap:        6,
			Padding:         16,
			Gap:             16,
			SelectionInset:  4,
		},
		Features: Features{
			ShowIconBadge: true,
			ShowLabels:    true,
		},
	}
}

func (t Theme) Normalize() Theme {
	return t
}

func (t Theme) Validate() error {
	if t.Window.Opacity < 1 {
		return fmt.Errorf("window.opacity must be between 1 and 255")
	}
	if t.Layout.ThumbnailWidth < 48 {
		return fmt.Errorf("layout.thumbnail_width must be >= 48")
	}
	if t.Layout.ThumbnailHeight < 40 {
		return fmt.Errorf("layout.thumbnail_height must be >= 40")
	}
	if t.Layout.IconSize < 16 {
		return fmt.Errorf("layout.icon_size must be >= 16")
	}
	if t.Layout.LabelHeight < 0 {
		return fmt.Errorf("layout.label_height must be >= 0")
	}
	if t.Layout.LabelGap < 0 {
		return fmt.Errorf("layout.label_gap must be >= 0")
	}
	if t.Layout.Padding < 0 {
		return fmt.Errorf("layout.padding must be >= 0")
	}
	if t.Layout.Gap < 0 {
		return fmt.Errorf("layout.gap must be >= 0")
	}
	if t.Layout.SelectionInset < 0 {
		return fmt.Errorf("layout.selection_inset must be >= 0")
	}
	return nil
}
