package config

type DevMode struct {
	Enabled         bool `toml:"-" env:"-"`
	ForceAgentLocal bool `toml:"force_agent_local" env:"HP_DEV_MODE_FORCE_AGENT_LOCAL"`
	// UserPassword, when set, becomes every user's password each time the app
	// starts: a dev server reseeded on each deploy keeps one password its team
	// knows, rather than the seed's, which is public.
	UserPassword string `toml:"user_password" env:"HP_DEV_MODE_USER_PASSWORD"`
	// LoggingQueryURL is where a backend run on the host reads the managed logs
	// backend. It queries it by service name on the stack's network, which a
	// process outside the swarm cannot resolve; `make local-logging-proxy`
	// forwards http://127.0.0.1:9428 to it.
	LoggingQueryURL string `toml:"logging_query_url" env:"HP_DEV_MODE_LOGGING_QUERY_URL"`
}
