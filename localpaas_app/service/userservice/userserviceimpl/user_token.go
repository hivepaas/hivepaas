package userserviceimpl

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity/appentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	MFATokenExp           = 60 * time.Second
	MFATotpSetupTokenExp  = 180 * time.Second
	UserInviteTokenExp    = 7 * timeutil.Day
	PasswordResetTokenExp = 7 * timeutil.Day
	ConsoleTokenExp       = 30 * time.Second
)

// GenerateMFAToken builds MFA token for using in the next step
func (s *service) GenerateMFAToken(userID string, mfaType base.MFAType, trustedDeviceID string) (string, error) {
	mfaToken, err := jwtsession.GenerateToken(&appentity.MFATokenClaims{
		Kind:            "mfa",
		UserID:          userID,
		MFAType:         mfaType,
		TrustedDeviceID: trustedDeviceID,
	}, MFATokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return mfaToken, nil
}

func (s *service) ParseMFAToken(token string) (*appentity.MFATokenClaims, error) {
	tokenClaims := &appentity.MFATokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "mfa" {
		return nil, apperrors.New(apperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateMFATotpSetupToken builds MFA TOTP token for setting up TOTP secret
func (s *service) GenerateMFATotpSetupToken(userID string, toptSecret string) (string, error) {
	mfaToken, err := jwtsession.GenerateToken(&appentity.MFATotpSetupTokenClaims{
		Kind:   "mfa-setup",
		UserID: userID,
		Secret: toptSecret,
	}, MFATotpSetupTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return mfaToken, nil
}

func (s *service) ParseMFATotpSetupToken(token string) (*appentity.MFATotpSetupTokenClaims, error) {
	tokenClaims := &appentity.MFATotpSetupTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "mfa-setup" {
		return nil, apperrors.New(apperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateUserInviteToken builds token for inviting users
func (s *service) GenerateUserInviteToken(userID string) (string, error) {
	token, err := jwtsession.GenerateToken(&appentity.UserInviteTokenClaims{
		Kind:   "user-invite",
		UserID: userID,
	}, UserInviteTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParseUserInviteToken(token string) (*appentity.UserInviteTokenClaims, error) {
	tokenClaims := &appentity.UserInviteTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "user-invite" {
		return nil, apperrors.New(apperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GeneratePasswordResetToken builds token for resetting passwords
func (s *service) GeneratePasswordResetToken(userID string) (string, error) {
	token, err := jwtsession.GenerateToken(&appentity.PasswordResetTokenClaims{
		Kind:   "pwd-reset",
		UserID: userID,
	}, PasswordResetTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParsePasswordResetToken(token string) (*appentity.PasswordResetTokenClaims, error) {
	tokenClaims := &appentity.PasswordResetTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "pwd-reset" {
		return nil, apperrors.New(apperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}

// GenerateConsoleToken builds token for accessing console or logs
func (s *service) GenerateConsoleToken(userID, targetID string) (string, error) {
	token, err := jwtsession.GenerateToken(&appentity.ConsoleTokenClaims{
		Kind:     "console",
		UserID:   userID,
		TargetID: targetID,
	}, ConsoleTokenExp)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}

func (s *service) ParseConsoleToken(token string) (*appentity.ConsoleTokenClaims, error) {
	tokenClaims := &appentity.ConsoleTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	if tokenClaims.Kind != "console" {
		return nil, apperrors.New(apperrors.ErrTokenInvalid)
	}
	return tokenClaims, nil
}
