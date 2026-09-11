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
	scope *entity.ObjectScope,
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

	if err = uc.authorizeReveal(ctx, db, auth, scope, setting); err != nil {
		return hperrors.Wrap(err)
	}
	if err = decrypter.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// RevealSubject identifies the secret being asked for.
//
// It exists because a stored setting is not the only secret the API hands out in
// the clear - the Swarm join token is another, and it is not a setting at all -
// and every one of them has to pass the same two gates and leave the same record.
type RevealSubject struct {
	Scope    base.ObjectScopeType
	ObjectID string
	Source   base.AuditLogSource

	// SecretType is the kind of secret being asked for, and decides whether the
	// operator's ReturnSecretsViaAPI flag can be stood down for it. Empty - which
	// is every stored setting - is always bound by the flag.
	SecretType base.SecretType

	ResType base.ResourceType
	ResID   string
	// ResName is stored as it reads now, because the object may be renamed or
	// deleted long before anyone comes to read the entry.
	ResName string

	// Detail is free-form JSON, and must never carry the secret it describes.
	Detail string
}

// AuthorizeSecretReveal decides whether the caller may see a secret in the clear,
// and records the answer either way.
//
// The record is written before the secret is handed over and its failure aborts
// the reveal. A secret released without a trace is the case this whole path
// exists to prevent, so a reveal that cannot be recorded does not happen.
//
// A refusal is recorded too, which is the half that matters: an admin passes the
// capability gate unconditionally, so the attempt is the only thing there is to
// see. Note this is the one place in the audit that records denials - everywhere
// else permission is settled in the handler, before the usecase runs.
func (uc *BaseUC) AuthorizeSecretReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	subject *RevealSubject,
) error {
	if subject == nil {
		return hperrors.NewArgumentInvalid("Reveal subject")
	}
	allowed, denyErr := uc.canRevealSecrets(ctx, db, auth, subject.SecretType)

	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}
	err := uc.AuditService.Record(ctx, db, &auditservice.Entry{
		Type:     base.AuditLogTypeSecretReveal,
		Scope:    subject.Scope,
		ObjectID: subject.ObjectID,
		Source:   subject.Source,
		Result:   result,
		Auth:     auth,
		ResType:  subject.ResType,
		ResID:    subject.ResID,
		ResName:  subject.ResName,
		Detail:   subject.Detail,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return denyErr
}

// authorizeReveal is AuthorizeSecretReveal for a stored setting.
func (uc *BaseUC) authorizeReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	setting *entity.Setting,
) error {
	return uc.AuthorizeSecretReveal(ctx, db, auth, &RevealSubject{
		Scope:    scope.ScopeType,
		ObjectID: scope.ScopeObjectID(),
		Source:   base.AuditLogSourceAPIGet,
		ResType:  base.ResourceTypeSetting,
		ResID:    setting.ID,
		ResName:  setting.Name,
	})
}

// canRevealSecrets reports whether the caller may see a secret in the clear.
//
// Two gates, and they answer different questions. The config flag is the
// operator's: it lives in a file on the host, and the one endpoint that can write
// it makes the caller re-enter the app secret, so a session that has taken over an
// admin account cannot turn it on. The capability is the account's. Note that an
// admin passes the second gate unconditionally - see CheckAccess - which is
// precisely why the attempt is recorded rather than merely refused.
//
// secretType only ever concerns the first gate. An operator can stand the flag
// down for a named kind of secret - see Security.AlwaysReturnSecretTypes - which
// is how node onboarding keeps working without opening up every stored
// credential. The capability is not exemptible: an exemption says this kind of
// secret is ordinary to hand out, not that anyone may have it.
//
// An empty secretType means a stored setting, and is always bound by the flag.
func (uc *BaseUC) canRevealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	secretType base.SecretType,
) (allowed bool, err error) {
	if !config.Current().Security.AllowsSecretType(secretType) {
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
