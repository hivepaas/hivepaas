package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

func InitConfig(lc fx.Lifecycle, cfg *config.Config, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			if err := validateConfig(cfg, logger); err != nil {
				return apperrors.New(err)
			}
			exportEnvVars(cfg, logger)
			exportDevEnvVars(cfg)
			if err := createSystemDataDirs(cfg, logger); err != nil {
				return apperrors.New(err)
			}

			// Register the channel to receive SIGHUP signal
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGHUP)

			// Start a goroutine to handle incoming signals
			go func() {
				for {
					sig := <-sigs // Block until a signal is received
					switch sig {
					case syscall.SIGHUP:
						logger.Info("SIGHUP received: Reloading configuration...")
						_, err := config.ReloadConfig()
						if err != nil {
							logger.Errorf("Failed to load configuration: %s", err)
						} else {
							logger.Info("SIGHUP handling: Configuration reloaded.")
						}
					default:
						// Do nothing
					}
				}
			}()
			return nil
		},
	})
}

func validateConfig(cfg *config.Config, logger logging.Logger) error {
	logger.Info("validating app config...")
	isProdEnv := !cfg.IsDevEnv()

	// JWT secret must not be empty or a trivial value
	if isProdEnv && len(cfg.Session.JWTSecret) < 10 {
		return fmt.Errorf("%w: invalid JWT secret for production", ErrInvalidConfig)
	}

	// Basic auth password must not be empty or a trivial value
	if isProdEnv && len(cfg.Session.BasicAuthPassword) < 10 {
		return fmt.Errorf("%w: basic auth password is invalid for production", ErrInvalidConfig)
	}

	return nil
}

func exportEnvVars(cfg *config.Config, logger logging.Logger) {
	logger.Info("exporting env variables...")

	if cfg.Proxy.HttpProxy != "" {
		_ = os.Setenv("HTTP_PROXY", cfg.Proxy.HttpProxy)
		_ = os.Setenv("http_proxy", cfg.Proxy.HttpProxy)
	}
	if cfg.Proxy.HttpsProxy != "" {
		_ = os.Setenv("HTTPS_PROXY", cfg.Proxy.HttpsProxy)
		_ = os.Setenv("https_proxy", cfg.Proxy.HttpsProxy)
	}
	if cfg.Proxy.NoProxy != "" {
		_ = os.Setenv("NO_PROXY", cfg.Proxy.NoProxy)
		_ = os.Setenv("no_proxy", cfg.Proxy.NoProxy)
	}
	if cfg.Proxy.Socks5Proxy != "" {
		_ = os.Setenv("SOCKS5_PROXY", cfg.Proxy.Socks5Proxy)
		_ = os.Setenv("socks5_proxy", cfg.Proxy.Socks5Proxy)
	}

	// TODO: More to export?
}

func exportDevEnvVars(cfg *config.Config) {
	if !cfg.IsLocalEnv() {
		return
	}

	// Add extra paths if running on Mac OS in order to make some commands work
	if runtime.GOOS == "darwin" {
		path := os.Getenv("PATH")
		extraPaths := []string{"/opt/homebrew/bin", "/usr/local/bin"}

		for _, p := range extraPaths {
			if !strings.Contains(path, p) {
				path = p + string(os.PathListSeparator) + path
			}
		}

		_ = os.Setenv("PATH", path)
	}
}

func createSystemDataDirs(cfg *config.Config, logger logging.Logger) error {
	logger.Info("creating data dirs if not exists...")

	for path, mode := range cfg.DataPathsToInitAtStartup() {
		err := os.MkdirAll(path, mode)
		if err != nil {
			logger.Errorf("failed to create data dir %s: %v", path, err)
			return apperrors.New(err)
		}
	}
	return nil
}
