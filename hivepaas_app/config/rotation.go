package config

import (
	"errors"
	"fmt"
)

var ErrSecretRotationUnavailable = errors.New("app secret rotation is unavailable")

// SaveAppSecret persists a new app secret.
//
// It is written only after the stored data encryption key has been rewrapped with
// it, so the file and the database never disagree in the direction that loses
// access: a crash before this leaves both still openable by the old secret.
func SaveAppSecret(newSecret string) error {
	if Current == nil {
		return fmt.Errorf("%w: config is not loaded", ErrSecretRotationUnavailable)
	}
	if newSecret == "" {
		return fmt.Errorf("%w: the new secret is empty", ErrSecretRotationUnavailable)
	}

	if err := saveManagedSettings(Current.AppPath, &ManagedSettings{Secret: newSecret}); err != nil {
		return fmt.Errorf("failed to persist the new app secret: %w", err)
	}

	Current.Secret = newSecret
	return nil
}
