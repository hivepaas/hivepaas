package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

// ListSSHKey Lists ssh-key settings
// @Summary Lists ssh-key settings
// @Description Lists ssh-key settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectSSHKey
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sshkeydto.ListSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys [get]
func (h *Handler) ListSSHKey(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}

// GetSSHKey Gets ssh-key setting details
// @Summary Gets ssh-key setting details
// @Description Gets ssh-key setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectSSHKey
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.GetSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys/{itemID} [get]
func (h *Handler) GetSSHKey(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}

// CreateSSHKey Creates a new ssh-key setting
// @Summary Creates a new ssh-key setting
// @Description Creates a new ssh-key setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectSSHKey
// @Param   projectID path string true "project ID"
// @Param   body body sshkeydto.CreateSSHKeyReq true "request data"
// @Success 201 {object} sshkeydto.CreateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys [post]
func (h *Handler) CreateSSHKey(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}

// UpdateSSHKey Updates ssh-key
// @Summary Updates ssh-key
// @Description Updates ssh-key
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSSHKey
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys/{itemID} [put]
func (h *Handler) UpdateSSHKey(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}

// UpdateSSHKeyStatus Updates ssh-key status
// @Summary Updates ssh-key status
// @Description Updates ssh-key status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSSHKeyStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyStatusReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys/{itemID}/status [put]
func (h *Handler) UpdateSSHKeyStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}

// DeleteSSHKey Deletes ssh-key setting
// @Summary Deletes ssh-key setting
// @Description Deletes ssh-key setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectSSHKey
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.DeleteSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssh-keys/{itemID} [delete]
func (h *Handler) DeleteSSHKey(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProject)
}
