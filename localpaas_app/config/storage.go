package config

type Storage struct {
	BindSource string `toml:"bind_source" env:"LP_STORAGE_BIND_SOURCE"`
	Volume     string `toml:"volume" env:"LP_STORAGE_VOLUME"`
}
