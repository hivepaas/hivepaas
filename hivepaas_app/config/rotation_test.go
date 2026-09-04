package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveAppSecret(t *testing.T) {
	t.Run("persists the new secret", func(t *testing.T) {
		Current = &Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "old-secret"}

		assert.NoError(t, SaveAppSecret("new-secret"))
		assert.Equal(t, "new-secret", Current.Secret)

		saved, err := loadManagedSettings(Current.AppPath)
		assert.NoError(t, err)
		assert.Equal(t, "new-secret", saved.Secret)
	})

	t.Run("refuses an empty secret", func(t *testing.T) {
		Current = &Config{Env: EnvDev, AppPath: t.TempDir(), Secret: "old-secret"}

		assert.ErrorIs(t, SaveAppSecret(""), ErrSecretRotationUnavailable)
		assert.Equal(t, "old-secret", Current.Secret)
	})
}
