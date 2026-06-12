package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"

	"github.com/localpaas/localpaas/localpaas_app/pkg/tracerr"
)

const (
	configFileName = "config.toml"
	envPrefix      = "LP"
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
)

const (
	RunModeApp          = "app"
	RunModeWorker       = "worker"
	RunModeAppAndWorker = "app+worker"
	RunModeUpdater      = "updater"
)

var (
	Current        *Config
	lastConfigFile string
)

type Config struct {
	Env      string `toml:"env" env:"LP_ENV"`
	Platform string `toml:"platform" env:"LP_PLATFORM" default:"remote"`
	RunMode  string `toml:"run_mode" env:"LP_RUN_MODE" default:"app+worker"`

	RootDomain string `toml:"root_domain" env:"LP_ROOT_DOMAIN"`
	AppDomain  string `toml:"app_domain" env:"LP_APP_DOMAIN"`
	BaseURL    string `toml:"base_url" env:"LP_APP_BASE_URL"`
	Secret     string `toml:"secret" env:"LP_APP_SECRET" default:"abc123"`
	AppPath    string `toml:"app_path" env:"LP_APP_PATH" default:"/var/lib/localpaas"`

	Users      Users      `toml:"users"`
	HTTPServer HTTPServer `toml:"http_server"`
	Storage    Storage    `toml:"storage"`
	DB         DB         `toml:"db"`
	Cache      Cache      `toml:"cache"`
	Session    Session    `toml:"session"`
	Proxy      Proxy      `toml:"proxy"`
	Tasks      Tasks      `toml:"tasks"`
	Files      Files      `toml:"files"`

	// Readonly
	SystemInfo SystemInfo `toml:"-"`
}

func (cfg *Config) IsDevEnv() bool   { return cfg.Env == EnvDev }
func (cfg *Config) IsLocalEnv() bool { return cfg.Platform == PlatformLocal }
func (cfg *Config) IsBetaEnv() bool  { return cfg.Env == EnvBeta }
func (cfg *Config) IsProdEnv() bool  { return cfg.Env == EnvProd }

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

func loadConfig(configFile string) (*Config, error) {
	config := &Config{}

	if configFile == "" {
		appPath := os.Getenv("LP_APP_PATH")
		if appPath == "" {
			appPath = "/var/lib/localpaas"
		}
		configFile = filepath.Join(appPath, configFileName)

		if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
			configFile = os.Getenv("LP_CONFIG_FILE")
			if configFile == "" {
				return nil, fmt.Errorf("%w: LP_CONFIG_FILE must be defined", ErrConfigFileUnset)
			}
		}
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrConfigFileNotFound, configFile)
	}

	err := configor.New(&configor.Config{ENVPrefix: envPrefix}).Load(config, configFile)
	if err != nil {
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
