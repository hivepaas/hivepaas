package config

type Cache struct {
	URL          string `toml:"url" env:"HP_CACHE_URL"`
	PoolSize     int    `toml:"pool_size" env:"HP_CACHE_POOL_SIZE" default:"10"`
	ReadTimeout  int    `toml:"read_timeout" env:"HP_CACHE_READ_TIMEOUT"`
	WriteTimeout int    `toml:"write_timeout" env:"HP_CACHE_WRITE_TIMEOUT"`
	MinIdleConns int    `toml:"min_idle_conns" env:"HP_CACHE_MIN_IDLE_CONNS"`
	UseTLS       bool   `toml:"use_tls" env:"HP_CACHE_USE_TLS" default:"false"`
}
