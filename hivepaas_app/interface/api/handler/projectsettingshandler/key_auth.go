package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/keyauthuc/keyauthdto"
)

// ListKeyAuth Lists key auth settings
// @Summary Lists key auth settings
// @Description Lists key auth settings
// @Tags    Project settings
// @Produce json
// @Id      listProjectKeyAuth
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} keyauthdto.ListKeyAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth [get]
func (h *Handler) ListKeyAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}

// GetKeyAuth Gets key auth setting details
// @Summary Gets key auth setting details
// @Description Gets key auth setting details
// @Tags    Project settings
// @Produce json
// @Id      getProjectKeyAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} keyauthdto.GetKeyAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth/{itemID} [get]
func (h *Handler) GetKeyAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}

// CreateKeyAuth Creates a new key auth setting
// @Summary Creates a new key auth setting
// @Description Creates a new key auth setting
// @Tags    Project settings
// @Produce json
// @Id      createProjectKeyAuth
// @Param   projectID path string true "project ID"
// @Param   body body keyauthdto.CreateKeyAuthReq true "request data"
// @Success 201 {object} keyauthdto.CreateKeyAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth [post]
func (h *Handler) CreateKeyAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}

// UpdateKeyAuth Updates key auth
// @Summary Updates key auth
// @Description Updates key auth
// @Tags    Project settings
// @Produce json
// @Id      updateProjectKeyAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body keyauthdto.UpdateKeyAuthReq true "request data"
// @Success 200 {object} keyauthdto.UpdateKeyAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth/{itemID} [put]
func (h *Handler) UpdateKeyAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}

// UpdateKeyAuthStatus Updates key auth status
// @Summary Updates key auth status
// @Description Updates key auth status
// @Tags    Project settings
// @Produce json
// @Id      updateProjectKeyAuthStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body keyauthdto.UpdateKeyAuthStatusReq true "request data"
// @Success 200 {object} keyauthdto.UpdateKeyAuthStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth/{itemID}/status [put]
func (h *Handler) UpdateKeyAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}

// DeleteKeyAuth Deletes key auth setting
// @Summary Deletes key auth setting
// @Description Deletes key auth setting
// @Tags    Project settings
// @Produce json
// @Id      deleteProjectKeyAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} keyauthdto.DeleteKeyAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/key-auth/{itemID} [delete]
func (h *Handler) DeleteKeyAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeKeyAuth, base.ObjectScopeProject)
}
