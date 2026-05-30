package fileserviceimpl

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity/appentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

func (s *service) GenerateDownloadToken(
	userID string,
	fileID string,
	requireLogin bool,
	expiration time.Duration,
) (string, error) {
	fileToken, err := jwtsession.GenerateToken(&appentity.FileDownloadTokenClaims{
		UserID:       userID,
		FileID:       fileID,
		RequireLogin: requireLogin,
	}, expiration)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return fileToken, nil
}

func (s *service) ParseDownloadToken(token string) (*appentity.FileDownloadTokenClaims, error) {
	tokenClaims := &appentity.FileDownloadTokenClaims{}
	if err := jwtsession.ParseToken(token, tokenClaims); err != nil {
		return nil, apperrors.New(apperrors.ErrTokenInvalid).WithCause(err)
	}
	return tokenClaims, nil
}
