package config

type DevMode struct {
	Enabled         bool `toml:"-" env:"-"`
	ForceAgentLocal bool `toml:"force_agent_local" env:"HP_DEV_MODE_FORCE_AGENT_LOCAL"`
	// UserPassword, when set, becomes every user's password each time the app
	// starts: a dev server reseeded on each deploy keeps one password its team
	// knows, rather than the seed's, which is public.
	UserPassword string `toml:"user_password" env:"HP_DEV_MODE_USER_PASSWORD"`
}
