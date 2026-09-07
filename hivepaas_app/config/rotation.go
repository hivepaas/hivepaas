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
	cfg := Current()
	if cfg == nil {
		return fmt.Errorf("%w: config is not loaded", ErrSecretRotationUnavailable)
	}
	if newSecret == "" {
		return fmt.Errorf("%w: the new secret is empty", ErrSecretRotationUnavailable)
	}

	err := updateManagedSettings(cfg.AppPath, func(settings *ManagedSettings) {
		settings.Secret = newSecret
	})
	if err != nil {
		return fmt.Errorf("failed to persist the new app secret: %w", err)
	}

	// Published as a copy, not written through the live pointer: readers hold that
	// pointer and must not have a field change under them mid-request. The copy is
	// shallow, which is safe because nothing here mutates the slices and maps it
	// shares - they are only ever replaced wholesale by a reload.
	updated := *cfg
	updated.Secret = newSecret
	SetCurrent(&updated)

	return nil
}
