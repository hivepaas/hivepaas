package projecthandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

// ListCloudStorage Lists cloud storages
// @Summary Lists cloud storages
// @Description Lists cloud storages
// @Tags    project_settings
// @Produce json
// @Id      listProjectCloudStorage
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cloudstoragedto.ListCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages [get]
func (h *Handler) ListCloudStorage(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}

// GetCloudStorage Gets cloud storage details
// @Summary Gets cloud storage details
// @Description Gets cloud storage details
// @Tags    project_settings
// @Produce json
// @Id      getProjectCloudStorage
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.GetCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages/{itemID} [get]
func (h *Handler) GetCloudStorage(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}

// CreateCloudStorage Creates a new cloud storage
// @Summary Creates a new cloud storage
// @Description Creates a new cloud storage
// @Tags    project_settings
// @Produce json
// @Id      createProjectCloudStorage
// @Param   projectID path string true "project ID"
// @Param   body body cloudstoragedto.CreateCloudStorageReq true "request data"
// @Success 201 {object} cloudstoragedto.CreateCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages [post]
func (h *Handler) CreateCloudStorage(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}

// UpdateCloudStorage Updates cloud storage
// @Summary Updates cloud storage
// @Description Updates cloud storage
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCloudStorage
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages/{itemID} [put]
func (h *Handler) UpdateCloudStorage(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}

// UpdateCloudStorageStatus Updates cloud storage status
// @Summary Updates cloud storage status
// @Description Updates cloud storage status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCloudStorageStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageStatusReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages/{itemID}/status [put]
func (h *Handler) UpdateCloudStorageStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}

// DeleteCloudStorage Deletes a cloud storage
// @Summary Deletes a cloud storage
// @Description Deletes a cloud storage
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectCloudStorage
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.DeleteCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cloud-storages/{itemID} [delete]
func (h *Handler) DeleteCloudStorage(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCloudStorage, base.SettingScopeProject)
}
