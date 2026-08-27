package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

// ListCloudStorage Lists cloud storages
// @Summary Lists cloud storages
// @Description Lists cloud storages
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvCloudStorage
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cloudstoragedto.ListCloudStorageResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages [get]
func (h *Handler) ListCloudStorage(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}

// GetCloudStorage Gets cloud storage details
// @Summary Gets cloud storage details
// @Description Gets cloud storage details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvCloudStorage
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.GetCloudStorageResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages/{itemID} [get]
func (h *Handler) GetCloudStorage(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}

// CreateCloudStorage Creates a new cloud storage
// @Summary Creates a new cloud storage
// @Description Creates a new cloud storage
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvCloudStorage
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body cloudstoragedto.CreateCloudStorageReq true "request data"
// @Success 201 {object} cloudstoragedto.CreateCloudStorageResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages [post]
func (h *Handler) CreateCloudStorage(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}

// UpdateCloudStorage Updates cloud storage
// @Summary Updates cloud storage
// @Description Updates cloud storage
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCloudStorage
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages/{itemID} [put]
func (h *Handler) UpdateCloudStorage(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}

// UpdateCloudStorageStatus Updates cloud storage status
// @Summary Updates cloud storage status
// @Description Updates cloud storage status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCloudStorageStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageStatusReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages/{itemID}/status [put]
func (h *Handler) UpdateCloudStorageStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}

// DeleteCloudStorage Deletes a cloud storage
// @Summary Deletes a cloud storage
// @Description Deletes a cloud storage
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvCloudStorage
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.DeleteCloudStorageResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/cloud-storages/{itemID} [delete]
func (h *Handler) DeleteCloudStorage(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeProjectEnv)
}
