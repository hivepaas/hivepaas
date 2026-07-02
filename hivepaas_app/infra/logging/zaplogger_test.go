package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

func TestNewZapLogger(t *testing.T) {
	t.Run("development", func(t *testing.T) {
		cfg := &config.Config{Env: config.EnvDev}
		logger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("production", func(t *testing.T) {
		cfg := &config.Config{Env: config.EnvProd}
		logger, err := NewZapLogger(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})
}

func TestZapLogger_Methods(t *testing.T) {
	cfg := &config.Config{Env: config.EnvDev}
	logger, _ := NewZapLogger(cfg)

	// Verify that calling these doesn't panic
	logger.Info("info", "key", "val")
	logger.Error("error", "key", "val")
	logger.Debug("debug", "key", "val")
	logger.Warn("warn", "key", "val")
	logger.Infof("infof %s", "arg")
	logger.Errorf("errorf %s", "arg")
	logger.Warnf("warnf %s", "arg")
	logger.Debugf("debugf %s", "arg")

	assert.NotNil(t, logger.(*ZapLogger).Sync)
}

func TestZapLogger_Panic(t *testing.T) {
	cfg := &config.Config{Env: config.EnvDev}
	logger, _ := NewZapLogger(cfg)

	assert.Panics(t, func() {
		logger.Panic("panic message")
	})

	assert.Panics(t, func() {
		logger.Panicf("panicf %s", "message")
	})
}
