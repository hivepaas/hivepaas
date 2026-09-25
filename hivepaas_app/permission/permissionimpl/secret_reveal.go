package permissionimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

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
func (p *manager) AuthorizeSecretReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	subject *permission.RevealSubject,
) error {
	if subject == nil {
		return hperrors.NewArgumentInvalid("Reveal subject")
	}
	allowed, denyErr := p.canRevealSecrets(ctx, db, auth, subject.SecretType)

	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}
	err := p.auditService.Record(ctx, db, &auditservice.Entry{
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
func (p *manager) canRevealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	secretType base.SecretType,
) (allowed bool, err error) {
	if !config.Current().Security.AllowsSecretType(secretType) {
		return false, hperrors.Wrap(hperrors.ErrRevealSecretsDisabled)
	}

	hasPerm, err := p.HasCapability(ctx, db, auth, base.ResourceCapSecretReveal)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if !hasPerm {
		return false, hperrors.Wrap(hperrors.ErrUserNotHavePermissionOnRevealSecrets)
	}
	return true, nil
}

func (p *manager) MayRevealSecrets(ctx context.Context, db database.IDB, auth *basedto.Auth) (bool, error) {
	allowed, err := p.canRevealSecrets(ctx, db, auth, "")
	if errors.Is(err, hperrors.ErrRevealSecretsDisabled) ||
		errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets) {
		return false, nil
	}
	return allowed, err
}
