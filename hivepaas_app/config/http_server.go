package config

import (
	"fmt"
)

type HTTPServer struct {
	BasePath         string   `toml:"base_path" env:"HP_HTTP_SERVER_BASE_PATH" default:"/_"`
	Port             int      `toml:"port" env:"HP_HTTP_SERVER_PORT" default:"10000"`
	CORSAllowOrigins []string `toml:"cors_allow_origins" env:"HP_HTTP_SERVER_CORS_ALLOW_ORIGINS"`

	// TrustedProxies are the networks whose X-Forwarded-For header is believed,
	// as CIDRs or plain addresses.
	//
	// It is empty by default, and empty means trust nobody: the client address is
	// then the address that actually opened the connection. Gin's own default is
	// the opposite - it trusts 0.0.0.0/0 - which means any caller can choose the
	// address that gets recorded against them simply by sending the header. That
	// is fine while nothing reads the address and useless the moment something
	// does, which is why this exists.
	//
	// Set it to the address of the proxy in front of this server, and nothing else.
	TrustedProxies []string `toml:"trusted_proxies" env:"HP_HTTP_SERVER_TRUSTED_PROXIES"`
}

func (c *HTTPServer) BindingAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}
