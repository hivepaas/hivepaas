package settings

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// secretDecrypter is implemented by every setting type that stores secrets.
//
// Reaching the decrypt through this interface rather than through each type's own
// MustAsX() is what lets one place cover all of them - and what makes a setting
// type added later covered without anybody remembering to come back here.
type secretDecrypter interface {
	Decrypt() error
}

// revealSecrets hands out the setting's secrets in the clear, if the caller asked
// for them and may have them.
//
// This lives here, once, rather than in each setting's own GetX. Repeated per
// type it is a check that fails open: forgetting it on one of fifteen usecases
// publishes that setting's secrets, silently, and nothing about the code looks
// wrong. Here, forgetting is not an option a future setting type has.
func (uc *BaseUC) revealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	reveal bool,
	setting *entity.Setting,
) error {
	if !reveal || setting == nil {
		return nil
	}
	// An inherited setting is read through, not owned, by this scope; its secrets
	// belong to whoever defined it.
	if setting.ObjectID != setting.CurrentObjectID {
		return nil
	}

	settingData, err := setting.Parse()
	if err != nil {
		return hperrors.Wrap(err)
	}
	decrypter, ok := settingData.(secretDecrypter)
	if !ok {
		return nil // the type holds no secrets, so there is nothing to reveal
	}

	if err = uc.authorizeReveal(ctx, db, auth, setting); err != nil {
		return hperrors.Wrap(err)
	}
	if err = decrypter.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// authorizeReveal decides whether the caller may see the secrets, and records the
// answer either way.
//
// The record is written before the secret is handed over and its failure aborts
// the reveal. A secret released without a trace is the case this whole path
// exists to prevent, so a reveal that cannot be recorded does not happen.
func (uc *BaseUC) authorizeReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	setting *entity.Setting,
) error {
	allowed, denyErr := uc.canRevealSecrets(ctx, db, auth)

	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}
	err := uc.AuditService.Record(ctx, db, &auditservice.Entry{
		Type:    base.AuditLogTypeSecretReveal,
		Source:  base.AuditLogSourceAPIGet,
		Result:  result,
		Auth:    auth,
		ResType: base.ResourceTypeSetting,
		ResID:   setting.ID,
		ResName: setting.Name,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return denyErr
}

// canRevealSecrets reports whether the caller may see stored secrets in the clear.
//
// Two gates, and they answer different questions. The config flag is the
// operator's: it lives in a file on the host, so a session that has taken over an
// admin account cannot turn it on. The capability is the account's. Note that an
// admin passes the second gate unconditionally - see CheckAccess - which is
// precisely why the attempt is recorded rather than merely refused.
func (uc *BaseUC) canRevealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
) (allowed bool, err error) {
	if !config.Current.Security.ReturnSecretsViaAPI {
		return false, hperrors.Wrap(hperrors.ErrRevealSecretsDisabled)
	}

	hasPerm, err := uc.HasCapability(ctx, db, auth, base.ResourceCapSecretReveal)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if !hasPerm {
		return false, hperrors.Wrap(hperrors.ErrUserNotHavePermissionOnRevealSecrets)
	}
	return true, nil
}

// HasCapability reports whether the caller holds a capability.
//
// It answers only that: the caller picks the error a refusal turns into, because
// "you may not reveal secrets" and "you may not mint API keys" are different
// things to tell someone even though the lookup behind them is the same.
func (uc *BaseUC) HasCapability(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	capability base.ResourceCapability,
) (bool, error) {
	hasPerm, err := uc.PermissionManager.CheckAccess(ctx, db, auth, &permission.CapabilityCheck{
		Capability: capability,
	})
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return hasPerm, nil
}
