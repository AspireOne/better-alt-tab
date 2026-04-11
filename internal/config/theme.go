package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"better_alt_tab/internal/theme"

	"github.com/BurntSushi/toml"
)

const themeDirName = "themes"

type themeFile struct {
	Version  int               `toml:"version"`
	Window   themeWindowFile   `toml:"window"`
	Colors   themeColorsFile   `toml:"colors"`
	Layout   themeLayoutFile   `toml:"layout"`
	Features themeFeaturesFile `toml:"features"`
}

type themeWindowFile struct {
	Opacity *int `toml:"opacity"`
}

type themeColorsFile struct {
	OverlayBackground           *string `toml:"overlay_background"`
	ItemBackground              *string `toml:"item_background"`
	ItemSelectedBackground      *string `toml:"item_selected_background"`
	ThumbnailFallbackBackground *string `toml:"thumbnail_fallback_background"`
	IconBadgeBackground         *string `toml:"icon_badge_background"`
	Label                       *string `toml:"label"`
	LabelSelected               *string `toml:"label_selected"`
}

type themeLayoutFile struct {
	ThumbnailWidth  *int32 `toml:"thumbnail_width"`
	ThumbnailHeight *int32 `toml:"thumbnail_height"`
	IconSize        *int32 `toml:"icon_size"`
	LabelHeight     *int32 `toml:"label_height"`
	LabelGap        *int32 `toml:"label_gap"`
	Padding         *int32 `toml:"padding"`
	Gap             *int32 `toml:"gap"`
	SelectionInset  *int32 `toml:"selection_inset"`
}

type themeFeaturesFile struct {
	ShowIconBadge *bool `toml:"show_icon_badge"`
	ShowLabels    *bool `toml:"show_labels"`
}

func ThemePath(name string) (string, error) {
	themeName, err := normalizeThemeName(name)
	if err != nil {
		return "", err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, configSubdir, appDirName, themeDirName, themeName+".toml"), nil
}

func LoadTheme(name string) (theme.Theme, error) {
	path, err := ThemePath(name)
	if err != nil {
		return theme.Default(), err
	}
	return loadThemePath(path)
}

func loadThemePath(path string) (theme.Theme, error) {
	current := theme.Default()

	// #nosec G304 -- The theme path comes from the app's fixed location or test-controlled temp files.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := saveThemePath(path, current); err != nil {
				return current, fmt.Errorf("create default theme %s: %w", path, err)
			}
			return current, nil
		}
		return current, fmt.Errorf("read theme %s: %w", path, err)
	}

	var decoded themeFile
	meta, err := toml.Decode(string(data), &decoded)
	if err != nil {
		return theme.Default(), fmt.Errorf("decode theme %s: %w", path, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, 0, len(undecoded))
		for _, key := range undecoded {
			keys = append(keys, key.String())
		}
		return theme.Default(), fmt.Errorf("decode theme %s: unknown keys: %s", path, strings.Join(keys, ", "))
	}

	if decoded.Version == 0 {
		decoded.Version = 1
	}
	if decoded.Version != 1 {
		return theme.Default(), fmt.Errorf("decode theme %s: unsupported version %d", path, decoded.Version)
	}

	if err := applyThemeFile(&current, decoded); err != nil {
		return theme.Default(), fmt.Errorf("decode theme %s: %w", path, err)
	}
	if err := current.Validate(); err != nil {
		return theme.Default(), fmt.Errorf("decode theme %s: %w", path, err)
	}

	if err := saveThemePath(path, current); err != nil {
		return current, fmt.Errorf("normalize theme %s: %w", path, err)
	}
	return current, nil
}

func saveThemePath(path string, current theme.Theme) error {
	if err := current.Validate(); err != nil {
		return fmt.Errorf("validate theme %s: %w", path, err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create theme directory %s: %w", dir, err)
	}

	spec := themeFile{
		Version: 1,
		Window: themeWindowFile{
			Opacity: intPtr(int(current.Window.Opacity)),
		},
		Colors: themeColorsFile{
			OverlayBackground:           stringPtr(formatColor(current.Colors.OverlayBackground)),
			ItemBackground:              stringPtr(formatColor(current.Colors.ItemBackground)),
			ItemSelectedBackground:      stringPtr(formatColor(current.Colors.ItemSelectedBackground)),
			ThumbnailFallbackBackground: stringPtr(formatColor(current.Colors.ThumbnailFallbackBackground)),
			IconBadgeBackground:         stringPtr(formatColor(current.Colors.IconBadgeBackground)),
			Label:                       stringPtr(formatColor(current.Colors.Label)),
			LabelSelected:               stringPtr(formatColor(current.Colors.LabelSelected)),
		},
		Layout: themeLayoutFile{
			ThumbnailWidth:  int32Ptr(current.Layout.ThumbnailWidth),
			ThumbnailHeight: int32Ptr(current.Layout.ThumbnailHeight),
			IconSize:        int32Ptr(current.Layout.IconSize),
			LabelHeight:     int32Ptr(current.Layout.LabelHeight),
			LabelGap:        int32Ptr(current.Layout.LabelGap),
			Padding:         int32Ptr(current.Layout.Padding),
			Gap:             int32Ptr(current.Layout.Gap),
			SelectionInset:  int32Ptr(current.Layout.SelectionInset),
		},
		Features: themeFeaturesFile{
			ShowIconBadge: boolPtr(current.Features.ShowIconBadge),
			ShowLabels:    boolPtr(current.Features.ShowLabels),
		},
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(spec); err != nil {
		return fmt.Errorf("encode theme %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write theme %s: %w", path, err)
	}
	return nil
}

func applyThemeFile(current *theme.Theme, decoded themeFile) error {
	if decoded.Window.Opacity != nil {
		if *decoded.Window.Opacity < 1 || *decoded.Window.Opacity > 255 {
			return fmt.Errorf("window.opacity must be between 1 and 255")
		}
		current.Window.Opacity = byte(*decoded.Window.Opacity)
	}

	if err := applyColor(&current.Colors.OverlayBackground, decoded.Colors.OverlayBackground); err != nil {
		return fmt.Errorf("colors.overlay_background: %w", err)
	}
	if err := applyColor(&current.Colors.ItemBackground, decoded.Colors.ItemBackground); err != nil {
		return fmt.Errorf("colors.item_background: %w", err)
	}
	if err := applyColor(&current.Colors.ItemSelectedBackground, decoded.Colors.ItemSelectedBackground); err != nil {
		return fmt.Errorf("colors.item_selected_background: %w", err)
	}
	if err := applyColor(&current.Colors.ThumbnailFallbackBackground, decoded.Colors.ThumbnailFallbackBackground); err != nil {
		return fmt.Errorf("colors.thumbnail_fallback_background: %w", err)
	}
	if err := applyColor(&current.Colors.IconBadgeBackground, decoded.Colors.IconBadgeBackground); err != nil {
		return fmt.Errorf("colors.icon_badge_background: %w", err)
	}
	if err := applyColor(&current.Colors.Label, decoded.Colors.Label); err != nil {
		return fmt.Errorf("colors.label: %w", err)
	}
	if err := applyColor(&current.Colors.LabelSelected, decoded.Colors.LabelSelected); err != nil {
		return fmt.Errorf("colors.label_selected: %w", err)
	}

	if decoded.Layout.ThumbnailWidth != nil {
		current.Layout.ThumbnailWidth = *decoded.Layout.ThumbnailWidth
	}
	if decoded.Layout.ThumbnailHeight != nil {
		current.Layout.ThumbnailHeight = *decoded.Layout.ThumbnailHeight
	}
	if decoded.Layout.IconSize != nil {
		current.Layout.IconSize = *decoded.Layout.IconSize
	}
	if decoded.Layout.LabelHeight != nil {
		current.Layout.LabelHeight = *decoded.Layout.LabelHeight
	}
	if decoded.Layout.LabelGap != nil {
		current.Layout.LabelGap = *decoded.Layout.LabelGap
	}
	if decoded.Layout.Padding != nil {
		current.Layout.Padding = *decoded.Layout.Padding
	}
	if decoded.Layout.Gap != nil {
		current.Layout.Gap = *decoded.Layout.Gap
	}
	if decoded.Layout.SelectionInset != nil {
		current.Layout.SelectionInset = *decoded.Layout.SelectionInset
	}

	if decoded.Features.ShowIconBadge != nil {
		current.Features.ShowIconBadge = *decoded.Features.ShowIconBadge
	}
	if decoded.Features.ShowLabels != nil {
		current.Features.ShowLabels = *decoded.Features.ShowLabels
	}

	return nil
}

func applyColor(dst *uint32, raw *string) error {
	if raw == nil {
		return nil
	}
	value, err := parseColor(*raw)
	if err != nil {
		return err
	}
	*dst = value
	return nil
}

func parseColor(raw string) (uint32, error) {
	if len(raw) != 7 || raw[0] != '#' {
		return 0, fmt.Errorf("must use #RRGGBB")
	}

	var value uint32
	for _, ch := range raw[1:] {
		value <<= 4
		switch {
		case ch >= '0' && ch <= '9':
			value |= uint32(ch - '0')
		case ch >= 'a' && ch <= 'f':
			value |= uint32(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			value |= uint32(ch-'A') + 10
		default:
			return 0, fmt.Errorf("must use #RRGGBB")
		}
	}

	r := (value >> 16) & 0xff
	g := (value >> 8) & 0xff
	b := value & 0xff
	return (b << 16) | (g << 8) | r, nil
}

func formatColor(color uint32) string {
	r := color & 0xff
	g := (color >> 8) & 0xff
	b := (color >> 16) & 0xff
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func normalizeThemeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	for _, ch := range name {
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '-' || ch == '_':
		default:
			return "", fmt.Errorf("invalid theme name %q", name)
		}
	}
	return name, nil
}

func boolPtr(value bool) *bool {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
