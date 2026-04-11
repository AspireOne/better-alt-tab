package config

const (
	appDirName   = "quick-app-switcher"
	configName   = "config.toml"
	configSubdir = ".config"
)

type Config struct {
	ShowThumbnails bool `toml:"show_thumbnails"`
}

func Default() Config {
	return Config{
		ShowThumbnails: true,
	}
}
