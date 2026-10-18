package config

import "strings"

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
	// LicenseFeatures are license features the installation has whatever its
	// license, comma-separated - managed-databases: a dev server reseeded on each
	// deploy keeps no key to grant them, and its paid features are there to be
	// tried. Ignored where dev mode is off, outside development and local.
	LicenseFeatures string `toml:"license_features" env:"HP_DEV_MODE_LICENSE_FEATURES"`
}

// LicenseFeatureList is LicenseFeatures as names, none when dev mode is off.
func (d *DevMode) LicenseFeatureList() []string {
	if !d.Enabled {
		return nil
	}
	var features []string
	for _, feature := range strings.Split(d.LicenseFeatures, ",") {
		if feature = strings.TrimSpace(feature); feature != "" {
			features = append(features, feature)
		}
	}
	return features
}
