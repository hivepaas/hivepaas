package config

import (
	"fmt"
)

type HTTPServer struct {
	BasePath         string   `toml:"base_path" env:"HP_HTTP_SERVER_BASE_PATH" default:"/_"`
	Port             int      `toml:"port" env:"HP_HTTP_SERVER_PORT" default:"10000"`
	CORSAllowOrigins []string `toml:"cors_allow_origins" env:"HP_HTTP_SERVER_CORS_ALLOW_ORIGINS"`
}

func (c *HTTPServer) BindingAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}
