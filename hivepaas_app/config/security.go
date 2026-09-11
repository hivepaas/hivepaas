package config

import (
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

	// AlwaysReturnSecretTypes are the kinds of secret the API hands out whatever
	// ReturnSecretsViaAPI says.
	//
	// The flag above is one switch over everything the database holds, which is
	// the right shape for stored secrets and the wrong shape for a secret that is
	// the only way to perform an ordinary operation - onboarding a node needs the
	// Swarm join token, and an operator should not have to open up every stored
	// credential to allow it. Naming a type here exempts that one kind.
	//
	// It exempts the operator's switch and nothing else: the caller still needs
	// the reveal capability, and the attempt is still recorded either way. Values
	// are base.SecretType; see base.AllSecretTypes for what may be listed.
	AlwaysReturnSecretTypes []string `toml:"always_return_secret_types" env:"HP_SECURITY_ALWAYS_RETURN_SECRET_TYPES"`
}

// Equal reports whether two sets of settings say the same thing.
//
// Written out because Security stopped being comparable with == the moment it
// grew a slice, and the update endpoint leans on that comparison to decide
// whether there is anything to persist. Order is not significant: the field is a
// set of exemptions, and reordering it is not a change.
func (s *Security) Equal(other *Security) bool {
	if s == nil || other == nil {
		return s == other
	}
	if s.ReturnSecretsViaAPI != other.ReturnSecretsViaAPI {
		return false
	}
	if len(s.AlwaysReturnSecretTypes) != len(other.AlwaysReturnSecretTypes) {
		return false
	}
	mine := gofn.ToSet(s.AlwaysReturnSecretTypes)
	theirs := gofn.ToSet(other.AlwaysReturnSecretTypes)
	if len(mine) != len(theirs) {
		return false
	}
	for _, value := range mine {
		if !gofn.Contain(theirs, value) {
			return false
		}
	}
	return true
}

// AllowsSecretType reports whether a kind of secret may be returned.
//
// An unnamed type - which is every stored setting - is bound by the flag alone.
func (s *Security) AllowsSecretType(secretType base.SecretType) bool {
	if s.ReturnSecretsViaAPI {
		return true
	}
	if secretType == "" {
		return false
	}
	return gofn.Contain(s.AlwaysReturnSecretTypes, string(secretType))
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
		// Written even when empty, so clearing the exemptions is not mistaken for
		// never having set them - the same reason the bool above is a pointer.
		exemptions := append([]string{}, security.AlwaysReturnSecretTypes...)
		settings.Security.AlwaysReturnSecretTypes = &exemptions
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
