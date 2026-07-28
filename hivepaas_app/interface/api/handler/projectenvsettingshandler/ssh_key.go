package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

// ListSSHKey Lists ssh-key settings
// @Summary Lists ssh-key settings
// @Description Lists ssh-key settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvSSHKey
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sshkeydto.ListSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys [get]
func (h *Handler) ListSSHKey(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}

// GetSSHKey Gets ssh-key setting details
// @Summary Gets ssh-key setting details
// @Description Gets ssh-key setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvSSHKey
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.GetSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys/{itemID} [get]
func (h *Handler) GetSSHKey(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}

// CreateSSHKey Creates a new ssh-key setting
// @Summary Creates a new ssh-key setting
// @Description Creates a new ssh-key setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvSSHKey
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body sshkeydto.CreateSSHKeyReq true "request data"
// @Success 201 {object} sshkeydto.CreateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys [post]
func (h *Handler) CreateSSHKey(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}

// UpdateSSHKey Updates ssh-key
// @Summary Updates ssh-key
// @Description Updates ssh-key
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSSHKey
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys/{itemID} [put]
func (h *Handler) UpdateSSHKey(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}

// UpdateSSHKeyStatus Updates ssh-key status
// @Summary Updates ssh-key status
// @Description Updates ssh-key status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSSHKeyStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyStatusReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys/{itemID}/status [put]
func (h *Handler) UpdateSSHKeyStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}

// DeleteSSHKey Deletes ssh-key setting
// @Summary Deletes ssh-key setting
// @Description Deletes ssh-key setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvSSHKey
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.DeleteSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssh-keys/{itemID} [delete]
func (h *Handler) DeleteSSHKey(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSHKey, base.ObjectScopeProjectEnv)
}
