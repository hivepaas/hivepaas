package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

func TestValidateConfig(t *testing.T) {
	logger := logging.GlobalLogger()

	t.Run("outside development the app needs both secrets", func(t *testing.T) {
		cfg := &config.Config{Env: config.EnvBeta, RunMode: config.RunModeApp}
		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)

		cfg.Session.JWTSecret = "0123456789abcdef0123456789abcdef"
		assert.ErrorIs(t, validateConfig(cfg, logger), ErrInvalidConfig)

		cfg.Secret = "0123456789abcdef"
		assert.NoError(t, validateConfig(cfg, logger))
	})

	// The agent issues no sessions and decrypts nothing, so it is given neither.
	t.Run("the agent needs neither", func(t *testing.T) {
		cfg := &config.Config{Env: config.EnvBeta, RunMode: config.RunModeAgent}
		assert.NoError(t, validateConfig(cfg, logger))
	})
}
