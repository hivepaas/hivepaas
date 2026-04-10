package projecthandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
)

// ListEmail Lists email settings
// @Summary Lists email settings
// @Description Lists email settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectEmail
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} emaildto.ListEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails [get]
func (h *Handler) ListEmail(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}

// GetEmail Gets email setting details
// @Summary Gets email setting details
// @Description Gets email setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectEmail
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.GetEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails/{itemID} [get]
func (h *Handler) GetEmail(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}

// CreateEmail Creates a new email setting
// @Summary Creates a new email setting
// @Description Creates a new email setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectEmail
// @Param   projectID path string true "project ID"
// @Param   body body emaildto.CreateEmailReq true "request data"
// @Success 201 {object} emaildto.CreateEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails [post]
func (h *Handler) CreateEmail(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}

// UpdateEmail Updates email setting
// @Summary Updates email setting
// @Description Updates email setting
// @Tags    project_settings
// @Produce json
// @Id      updateProjectEmail
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails/{itemID} [put]
func (h *Handler) UpdateEmail(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}

// UpdateEmailStatus Updates Email status setting
// @Summary Updates Email status setting
// @Description Updates Email status setting
// @Tags    project_settings
// @Produce json
// @Id      updateProjectEmailStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailStatusReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails/{itemID}/status [put]
func (h *Handler) UpdateEmailStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}

// DeleteEmail Deletes email setting
// @Summary Deletes email setting
// @Description Deletes email setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectEmail
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.DeleteEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/emails/{itemID} [delete]
func (h *Handler) DeleteEmail(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeEmail, base.SettingScopeProject)
}
