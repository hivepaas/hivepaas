package useruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// The forms an account can be written through, as the entry's section. See
// AuditLogTypeUserUpdate.
const (
	// auditSectionInvite and auditSectionSignup are the two ways an account
	// appears: somebody with admin made it, or somebody made their own.
	auditSectionInvite = "invite"
	auditSectionSignup = "signup"

	// auditSectionAccount is the admin form - role, status, security option, and
	// the grants. auditSectionProfile is what an account may change about itself.
	auditSectionAccount = "account"
	auditSectionProfile = "profile"

	auditSectionPassword             = "password"
	auditSectionPasswordReset        = "password-reset"
	auditSectionPasswordResetRequest = "password-reset-request" //nolint:gosec // G101: a section name

	auditSectionMFASetup  = "mfa-setup"
	auditSectionMFARemove = "mfa-remove"
)

// Why a request against an account was refused.
const (
	auditReasonUnknownEmail  = "unknown-email"
	auditReasonWrongPasscode = "wrong-passcode"
	auditReasonNotAllowed    = "not-allowed"
)

// recordUserChange records something done to an account.
//
// The account acted on is the entry's object and its resource; who did it comes
// from auth. The two are the same for the self-service forms and different for
// the admin ones, which is most of what these entries are read for - "who
// changed whose role" has no other answer.
func (uc *UC) recordUserChange(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	logType base.AuditLogType,
	section string,
	user *entity.User,
	detail *auditdetail.Builder,
) error {
	if user == nil {
		return hperrors.NewArgumentInvalid("audited user")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     logType,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.ID,
		Source:   userAuditSource(logType),
		Section:  section,
		Auth:     auth,
		ResType:  base.ResourceTypeUser,
		ResID:    user.ID,
		ResName:  user.Username,
		Detail:   detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// recordUserChangeDenied records a refused attempt against an account.
//
// Separate from the allowed path because RecordAllowed refuses an entry it
// cannot attribute, and here that refusal is the wrong answer: a reset asked for
// an address nobody has is precisely an attempt with nobody behind it, and it is
// the one worth keeping.
//
// user is nil when the account could not be identified - and nothing the caller
// typed is written down either way. An email box holds whatever somebody pasted
// into it, and the reason plus the address the request came from is the whole of
// what the entry has to say. Same rule as the refused logins next door.
func (uc *UC) recordUserChangeDenied(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	section string,
	user *entity.User,
	reason string,
) error {
	entry := &auditservice.Entry{
		Type:    base.AuditLogTypeUserUpdate,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: section,
		Result:  base.AuditLogResultDenied,
		Auth:    auth,
		ResType: base.ResourceTypeUser,
		Detail:  auditdetail.New().Set("reason", reason).String(),
	}
	if user != nil {
		entry.Scope = base.ObjectScopeUser
		entry.ObjectID = user.ID
		entry.ResID = user.ID
		entry.ResName = user.Username
	}

	if err := uc.auditService.Record(ctx, db, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// userAuditSource keeps the entry's source in step with what happened to the
// account, so a created account does not read as an edit.
func userAuditSource(logType base.AuditLogType) base.AuditLogSource {
	switch logType { //nolint:exhaustive // only the three an account can be written by
	case base.AuditLogTypeUserCreate:
		return base.AuditLogSourceAPICreate
	case base.AuditLogTypeUserDelete:
		return base.AuditLogSourceAPIDelete
	}
	return base.AuditLogSourceAPIUpdate
}

// authOfUser is the account acting for itself, on the paths where a link rather
// than a session is the proof.
//
// A password reset token and an invite token are both possession of the address
// the account is reached at, which is what the account is. Recording those as
// nobody would throw away the only attribution there is; recording them as the
// user says what happened, and the section says which door it came through.
func authOfUser(user *entity.User) *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: user}}
}

// recordPasswordResetRequest records a reset link going out for an account,
// asked for by somebody the request cannot name.
//
// The forgot-password form takes an address and nothing else, so there is no
// actor - and attributing it to the account would have the entry say the account
// asked, which is exactly what it would say for an attacker making the mail go
// out. What is true is that a link was sent for this account, from this address,
// at this time, and that is what goes in.
//
// The admin-initiated reset next door is a different entry: that one has an
// actor, and it is recorded through recordUserChange like every other admin act.
func (uc *UC) recordPasswordResetRequest(
	ctx context.Context,
	db database.IDB,
	user *entity.User,
) error {
	if user == nil {
		return hperrors.NewArgumentInvalid("audited user")
	}

	err := uc.auditService.Record(ctx, db, &auditservice.Entry{
		Type:     base.AuditLogTypeUserUpdate,
		Scope:    base.ObjectScopeUser,
		ObjectID: user.ID,
		Source:   base.AuditLogSourceAPIUpdate,
		Section:  auditSectionPasswordResetRequest,
		Result:   base.AuditLogResultAllowed,
		ResType:  base.ResourceTypeUser,
		ResID:    user.ID,
		ResName:  user.Username,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
