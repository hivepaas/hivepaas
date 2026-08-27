package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
)

// ListEmail Lists email settings
// @Summary Lists email settings
// @Description Lists email settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvEmail
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} emaildto.ListEmailResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails [get]
func (h *Handler) ListEmail(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}

// GetEmail Gets email setting details
// @Summary Gets email setting details
// @Description Gets email setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvEmail
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.GetEmailResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails/{itemID} [get]
func (h *Handler) GetEmail(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}

// CreateEmail Creates a new email setting
// @Summary Creates a new email setting
// @Description Creates a new email setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvEmail
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body emaildto.CreateEmailReq true "request data"
// @Success 201 {object} emaildto.CreateEmailResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails [post]
func (h *Handler) CreateEmail(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}

// UpdateEmail Updates email setting
// @Summary Updates email setting
// @Description Updates email setting
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvEmail
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails/{itemID} [put]
func (h *Handler) UpdateEmail(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}

// UpdateEmailStatus Updates Email status setting
// @Summary Updates Email status setting
// @Description Updates Email status setting
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvEmailStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailStatusReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails/{itemID}/status [put]
func (h *Handler) UpdateEmailStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}

// DeleteEmail Deletes email setting
// @Summary Deletes email setting
// @Description Deletes email setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvEmail
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.DeleteEmailResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/emails/{itemID} [delete]
func (h *Handler) DeleteEmail(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeProjectEnv)
}
