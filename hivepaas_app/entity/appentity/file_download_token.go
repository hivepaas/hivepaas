package appentity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
)

type FileDownloadTokenClaims struct {
	jwtsession.BaseClaims
	FileID       string `json:"fileId"`
	UserID       string `json:"userId"`
	RequireLogin bool   `json:"requireLogin,omitempty"`
}
