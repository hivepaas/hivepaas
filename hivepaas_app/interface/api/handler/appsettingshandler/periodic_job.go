package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
)

// ListAppPeriodicJob Lists periodic jobs
// @Summary Lists periodic jobs
// @Description Lists periodic jobs
// @Tags    app_settings
// @Produce json
// @Id      listAppPeriodicJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} periodicjobdto.ListPeriodicJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs [get]
func (h *Handler) ListAppPeriodicJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}

// GetAppPeriodicJob Gets periodic job details
// @Summary Gets periodic job details
// @Description Gets periodic job details
// @Tags    app_settings
// @Produce json
// @Id      getAppPeriodicJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} periodicjobdto.GetPeriodicJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs/{itemID} [get]
func (h *Handler) GetAppPeriodicJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}

// CreateAppPeriodicJob Creates a new periodic job
// @Summary Creates a new periodic job
// @Description Creates a new periodic job
// @Tags    app_settings
// @Produce json
// @Id      createAppPeriodicJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body periodicjobdto.CreatePeriodicJobReq true "request data"
// @Success 201 {object} periodicjobdto.CreatePeriodicJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs [post]
func (h *Handler) CreateAppPeriodicJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}

// UpdateAppPeriodicJob Updates a periodic job
// @Summary Updates a periodic job
// @Description Updates a periodic job
// @Tags    app_settings
// @Produce json
// @Id      updateAppPeriodicJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body periodicjobdto.UpdatePeriodicJobReq true "request data"
// @Success 200 {object} periodicjobdto.UpdatePeriodicJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs/{itemID} [put]
func (h *Handler) UpdateAppPeriodicJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}

// UpdateAppPeriodicJobStatus Updates periodic job status
// @Summary Updates periodic job status
// @Description Updates periodic job status
// @Tags    app_settings
// @Produce json
// @Id      updateAppPeriodicJobStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body periodicjobdto.UpdatePeriodicJobStatusReq true "request data"
// @Success 200 {object} periodicjobdto.UpdatePeriodicJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs/{itemID}/status [put]
func (h *Handler) UpdateAppPeriodicJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}

// DeleteAppPeriodicJob Deletes periodic job
// @Summary Deletes periodic job
// @Description Deletes periodic job
// @Tags    app_settings
// @Produce json
// @Id      deleteAppPeriodicJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} periodicjobdto.DeletePeriodicJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/periodic-jobs/{itemID} [delete]
func (h *Handler) DeleteAppPeriodicJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypePeriodicJob, base.ObjectScopeApp)
}
