package config

type Storage struct {
	BindSource string `toml:"bind_source" env:"HP_STORAGE_BIND_SOURCE"`
	Volume     string `toml:"volume" env:"HP_STORAGE_VOLUME"`
}
