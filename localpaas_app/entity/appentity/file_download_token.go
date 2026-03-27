package appentity

import (
	"github.com/localpaas/localpaas/localpaas_app/pkg/jwtsession"
)

type FileDownloadTokenClaims struct {
	jwtsession.BaseClaims
	FileID       string `json:"fileID"`
	UserID       string `json:"userID"`
	RequireLogin bool   `json:"requireLogin,omitempty"`
}
