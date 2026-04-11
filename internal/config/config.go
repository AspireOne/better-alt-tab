package config

const (
	appDirName   = "quick-app-switcher"
	configName   = "config.toml"
	configSubdir = ".config"
)

type Config struct {
	ShowThumbnails       bool `toml:"show_thumbnails"`
	LaunchOnStartup      bool `toml:"launch_on_startup"`
	InstantSwitchPreview bool `toml:"instant_switch_preview"`
}

func Default() Config {
	return Config{
		ShowThumbnails:       true,
		LaunchOnStartup:      false,
		InstantSwitchPreview: true,
	}
}
