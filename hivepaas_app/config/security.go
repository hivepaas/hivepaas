package config

import (
	"errors"
	"fmt"
)

// Security holds the operator-level switches.
//
// NOTE when adding a field here: the update endpoint applies a change by asking
// the main app to re-read its config (hpappservice.ReloadHpAppConfig, a SIGHUP).
// The worker is a separate swarm service and does not get that signal, so a flag
// the worker acts on would take effect in the API immediately and in the worker
// only at its next restart. Add a worker reload before adding such a flag.
type Security struct {
	ReturnSecretsViaAPI bool `toml:"return_secrets_via_api" env:"HP_SECURITY_RETURN_SECRETS_VIA_API"`
}

var ErrSecuritySettingsUnavailable = errors.New("security settings cannot be saved")

// SaveSecuritySettings persists the security settings and applies them here.
//
// They go to the managed file rather than to the database for the same reason the
// app secret does: this is the operator's layer, not the account's. A setting the
// app can rewrite in its own database is a setting anyone who reaches the database
// can rewrite, and the flags here decide whether stored secrets may leave the
// server at all - see the reveal path in usecase/settings.
//
// Every field is written explicitly, including the ones that are false. Writing
// only the true ones would make "turn it off" a no-op, since an absent key means
// "leave whatever was loaded alone".
func SaveSecuritySettings(security *Security) error {
	cfg := Current()
	if cfg == nil {
		return fmt.Errorf("%w: config is not loaded", ErrSecuritySettingsUnavailable)
	}
	if security == nil {
		return fmt.Errorf("%w: no settings given", ErrSecuritySettingsUnavailable)
	}

	err := updateManagedSettings(cfg.AppPath, func(settings *ManagedSettings) {
		settings.Security.ReturnSecretsViaAPI = &security.ReturnSecretsViaAPI
	})
	if err != nil {
		return fmt.Errorf("failed to persist the security settings: %w", err)
	}

	// A copy, for the reason SaveAppSecret gives: the live pointer is held by
	// readers and its contents must not change under them.
	updated := *cfg
	updated.Security = *security
	SetCurrent(&updated)

	return nil
}
