package config

type Agent struct {
	Port        int    `toml:"port" env:"HP_AGENT_PORT" default:"10001"`
	SecretToken string `toml:"secret_token" env:"HP_AGENT_SECRET_TOKEN"`
}
