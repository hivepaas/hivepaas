package config

type DevMode struct {
	Enabled         bool `toml:"-" env:"-"`
	ForceAgentLocal bool `toml:"force_agent_local" env:"HP_DEV_MODE_FORCE_AGENT_LOCAL"`
}
