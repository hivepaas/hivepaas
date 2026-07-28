package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
)

// ListIMService Lists IM services
// @Summary Lists IM services
// @Description Lists IM services
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvIMService
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} imservicedto.ListIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services [get]
func (h *Handler) ListIMService(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}

// GetIMService Gets IM service details
// @Summary Gets IM service details
// @Description Gets IM service details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvIMService
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} imservicedto.GetIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services/{itemID} [get]
func (h *Handler) GetIMService(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}

// CreateIMService Creates a new IM service
// @Summary Creates a new IM service
// @Description Creates a new IM service
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvIMService
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body imservicedto.CreateIMServiceReq true "request data"
// @Success 201 {object} imservicedto.CreateIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services [post]
func (h *Handler) CreateIMService(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}

// UpdateIMService Updates IM service
// @Summary Updates IM service
// @Description Updates IM service
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvIMService
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body imservicedto.UpdateIMServiceReq true "request data"
// @Success 200 {object} imservicedto.UpdateIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services/{itemID} [put]
func (h *Handler) UpdateIMService(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}

// UpdateIMServiceStatus Updates IM service status
// @Summary Updates IM service status
// @Description Updates IM service status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvIMServiceStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body imservicedto.UpdateIMServiceStatusReq true "request data"
// @Success 200 {object} imservicedto.UpdateIMServiceStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services/{itemID}/status [put]
func (h *Handler) UpdateIMServiceStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}

// DeleteIMService Deletes IM service
// @Summary Deletes IM service
// @Description Deletes IM service
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvIMService
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} imservicedto.DeleteIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/im-services/{itemID} [delete]
func (h *Handler) DeleteIMService(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeProjectEnv)
}
