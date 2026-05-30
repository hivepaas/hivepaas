package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
)

// ListNotification Lists notification settings
// @Summary Lists notification settings
// @Description Lists notification settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectNotification
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} notificationdto.ListNotificationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications [get]
func (h *Handler) ListNotification(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}

// GetNotification Gets notification setting details
// @Summary Gets notification setting details
// @Description Gets notification setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectNotification
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} notificationdto.GetNotificationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications/{itemID} [get]
func (h *Handler) GetNotification(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}

// CreateNotification Creates a new notification setting
// @Summary Creates a new notification setting
// @Description Creates a new notification setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectNotification
// @Param   projectID path string true "project ID"
// @Param   body body notificationdto.CreateNotificationReq true "request data"
// @Success 201 {object} notificationdto.CreateNotificationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications [post]
func (h *Handler) CreateNotification(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}

// UpdateNotification Updates notification
// @Summary Updates notification
// @Description Updates notification
// @Tags    project_settings
// @Produce json
// @Id      updateProjectNotification
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body notificationdto.UpdateNotificationReq true "request data"
// @Success 200 {object} notificationdto.UpdateNotificationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications/{itemID} [put]
func (h *Handler) UpdateNotification(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}

// UpdateNotificationStatus Updates notification status
// @Summary Updates notification status
// @Description Updates notification status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectNotificationStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body notificationdto.UpdateNotificationStatusReq true "request data"
// @Success 200 {object} notificationdto.UpdateNotificationStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications/{itemID}/status [put]
func (h *Handler) UpdateNotificationStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}

// DeleteNotification Deletes notification setting
// @Summary Deletes notification setting
// @Description Deletes notification setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectNotification
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} notificationdto.DeleteNotificationResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/notifications/{itemID} [delete]
func (h *Handler) DeleteNotification(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeNotification, base.ObjectScopeProject)
}
