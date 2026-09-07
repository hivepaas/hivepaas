package config

type Security struct {
	ReturnSecretsViaAPI bool `toml:"return_secrets_via_api" env:"HP_SECURITY_RETURN_SECRETS_VIA_API" default:"true"`
}
