package config

type Proxy struct {
	HttpProxy   string `toml:"http_proxy" env:"HTTP_PROXY"`
	HttpsProxy  string `toml:"https_proxy" env:"HTTPS_PROXY"`
	NoProxy     string `toml:"no_proxy" env:"NO_PROXY"`
	Socks5Proxy string `toml:"socks5_proxy" env:"SOCKS5_PROXY"`
}
