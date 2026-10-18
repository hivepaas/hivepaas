package appmanageddto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type RefreshManagedCredentialsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewRefreshManagedCredentialsReq() *RefreshManagedCredentialsReq {
	return &RefreshManagedCredentialsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RefreshManagedCredentialsReq) Validate() hperrors.ValidationErrors {
	return validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID)
}

type RefreshManagedCredentialsResp struct {
	Meta *basedto.Meta                      `json:"meta"`
	Data *RefreshManagedCredentialsDataResp `json:"data"`
}

type RefreshManagedCredentialsDataResp struct {
	// Changed says the app's credentials changed and are being mounted again,
	// which restarts the database's task.
	Changed bool `json:"changed"`
}
