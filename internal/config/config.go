package config

const (
	appDirName   = "quick-app-switcher"
	configName   = "config.toml"
	configSubdir = ".config"
)

type Config struct {
	ShowThumbnails  bool `toml:"show_thumbnails"`
	LaunchOnStartup bool `toml:"launch_on_startup"`
}

func Default() Config {
	return Config{
		ShowThumbnails:  true,
		LaunchOnStartup: false,
	}
}
