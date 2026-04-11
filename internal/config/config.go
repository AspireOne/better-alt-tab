package config

const (
	appDirName   = "better-alt-tab"
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
