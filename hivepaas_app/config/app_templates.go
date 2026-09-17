package config

// AppTemplates configures where app templates come from.
type AppTemplates struct {
	// Dir reads templates from a checkout of the app-templates repository instead
	// of the signed official source, with no signature and no hashes. It is for
	// authoring templates, and it is honored only in the development environment.
	Dir string `toml:"dir" env:"HP_TEMPLATES_DIR"`
}
