package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

const (
	configFileName = "config.toml"
	defaultAppPath = "/var/lib/hivepaas"
	// envPrefix namespaces the env names configor derives on its own. Every scalar
	// setting here carries an explicit `env:"HP_..."` tag, which configor uses
	// verbatim, so the prefix only ever applies to the nested section structs
	// (HP_DB, HP_SESSION, ...) that can be fed a whole YAML blob.
	//
	// It cannot simply be dropped: with an empty ENVPrefix configor falls back to
	// "Configor" and starts honoring CONFIGOR_ENV_PREFIX, and with "-" it drops
	// the prefix entirely and would read bare names like DB or CACHE - both worse
	// than an explicit namespace of our own.
	envPrefix = "HP"
)

const (
	PlatformLocal  = "local"
	PlatformRemote = "remote"

	EnvDev  = "development"
	EnvBeta = "beta"
	EnvProd = "production"
)

var (
	ErrConfigFileUnset    = errors.New("config file unset")
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrAppSecretUnset     = errors.New("app secret is not configured")
)

const (
	RunModeApp          = "app"
	RunModeWorker       = "worker"
	RunModeAppAndWorker = "app+worker"
	RunModeAgent        = "agent"
	RunModeUpdater      = "updater"
)

var (
	Current        *Config
	lastConfigFile string
)

type Config struct {
	Env      string `toml:"env" env:"HP_ENV"`
	Platform string `toml:"platform" env:"HP_PLATFORM" default:"remote"`
	RunMode  string `toml:"run_mode" env:"HP_RUN_MODE" default:"app+worker"`

	RootDomain string `toml:"root_domain" env:"HP_ROOT_DOMAIN"`
	AppDomain  string `toml:"app_domain" env:"HP_APP_DOMAIN"`
	// Secret is the key every stored secret is encrypted with. It has no default
	// on purpose: a well-known one would mean every zero-config install shares a
	// publicly known key. See ensureAppSecret.
	Secret  string `toml:"secret" env:"HP_APP_SECRET"`
	AppPath string `toml:"app_path" env:"HP_APP_PATH" default:"/var/lib/hivepaas"`

	Users      Users      `toml:"users"`
	HTTPServer HTTPServer `toml:"http_server"`
	Storage    Storage    `toml:"storage"`
	DB         DB         `toml:"db"`
	Cache      Cache      `toml:"cache"`
	Session    Session    `toml:"session"`
	Proxy      Proxy      `toml:"proxy"`
	Tasks      Tasks      `toml:"tasks"`
	Files      Files      `toml:"files"`
	Agent      Agent      `toml:"agent"`

	DevMode DevMode `toml:"dev_mode"`

	// Readonly and internal data
	SystemInfo SystemInfo `toml:"-"`
}

func (cfg *Config) IsDevEnv() bool   { return cfg.Env == EnvDev }
func (cfg *Config) IsLocalEnv() bool { return cfg.Platform == PlatformLocal }
func (cfg *Config) IsBetaEnv() bool  { return cfg.Env == EnvBeta }
func (cfg *Config) IsProdEnv() bool  { return cfg.Env == EnvProd }

func (cfg *Config) BaseURL() string {
	if cfg.Platform == PlatformLocal {
		return fmt.Sprintf("http://%s:%v", cfg.loadAppDomain(), cfg.HTTPServer.Port)
	}
	return "https://" + cfg.loadAppDomain()
}

/// LOAD CONFIG

func LoadConfig() (*Config, error) {
	if Current != nil {
		return Current, nil
	}
	cfg, err := loadConfig("")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	Current = cfg
	return cfg, nil
}

// resolveAppPath returns the app directory, where both the config file and the
// managed settings live. It has to read the environment directly: the value is
// needed to find the very file the config is loaded from.
func resolveAppPath() string {
	if appPath := os.Getenv("HP_APP_PATH"); appPath != "" {
		return appPath
	}
	return defaultAppPath
}

func loadConfig(configFile string) (*Config, error) {
	config := &Config{}
	appPath := resolveAppPath()

	if configFile == "" {
		configFile = filepath.Join(appPath, configFileName)

		// #nosec G703
		if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
			configFile = os.Getenv("HP_CONFIG_FILE")
			if configFile == "" {
				return nil, fmt.Errorf("%w: HP_CONFIG_FILE must be defined", ErrConfigFileUnset)
			}
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) { // #nosec G703
		return nil, fmt.Errorf("%w: %s", ErrConfigFileNotFound, configFile)
	}

	// The environment is read from our own snapshot rather than from the process,
	// which is cleared below so user-supplied commands cannot read the secrets out
	// of it. Without this a reload would see an empty environment.
	snapshotEnv()
	restoreEnv()
	defer clearEnv()

	err := configor.New(&configor.Config{ENVPrefix: envPrefix}).Load(config, configFile)
	if err != nil {
		return config, tracerr.Wrap(err)
	}

	// Applied last so a setting the app rotates for itself wins over a stale value
	// still present in the environment or the base config file.
	managed, err := loadManagedSettings(appPath)
	if err != nil {
		return config, tracerr.Wrap(err)
	}
	managed.applyTo(config)

	// Turn on dev mode for dev/local env
	config.DevMode.Enabled = config.IsDevEnv() || config.IsLocalEnv()

	if err := ensureAppSecret(config, appPath); err != nil {
		return config, tracerr.Wrap(err)
	}

	lastConfigFile = configFile
	return config, nil
}

func ReloadConfig() (*Config, error) {
	newConfig, err := loadConfig(lastConfigFile)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// TODO: validate then apply a certain portion of the new config

	Current = newConfig
	return newConfig, nil
}

// appSecretLen is the length in bytes of a generated app secret, before hex
// encoding doubles it.
const appSecretLen = 32

// ensureAppSecret makes sure encryption is always on.
//
// Storing credentials unencrypted is not an option the operator gets: the most
// realistic way they leak is a database dump leaving the host - which this app
// does on purpose, backing up to cloud storage - and a dump is only useless to
// whoever finds it if the values in it are encrypted.
//
// Outside development the secret has to be configured explicitly. Generating one
// here would be unsafe with more than one replica: replicas do not share the app
// volume, so each would invent a different key and encrypt rows the others cannot
// read. Development gets a generated one so a fresh checkout just runs.
func ensureAppSecret(config *Config, appPath string) error {
	if config.Secret != "" {
		return nil
	}

	if !config.IsDevEnv() {
		return fmt.Errorf("%w: HP_APP_SECRET must be set, e.g. %s",
			ErrAppSecretUnset, gofn.RandTokenAsHex(appSecretLen))
	}

	config.Secret = gofn.RandTokenAsHex(appSecretLen)
	if err := saveManagedSettings(appPath, &ManagedSettings{Secret: config.Secret}); err != nil {
		return fmt.Errorf("failed to persist the generated app secret: %w", err)
	}
	return nil
}
