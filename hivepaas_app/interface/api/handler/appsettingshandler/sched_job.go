package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ListAppSchedJob Lists sched-jobs
// @Summary Lists sched-jobs
// @Description Lists sched-jobs
// @Tags    app_settings
// @Produce json
// @Id      listAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} schedjobdto.ListSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs [get]
func (h *Handler) ListAppSchedJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// GetAppSchedJob Gets sched-job details
// @Summary Gets sched-job details
// @Description Gets sched-job details
// @Tags    app_settings
// @Produce json
// @Id      getAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.GetSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs/{itemID} [get]
func (h *Handler) GetAppSchedJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// CreateAppSchedJob Creates a new sched-job
// @Summary Creates a new sched-job
// @Description Creates a new sched-job
// @Tags    app_settings
// @Produce json
// @Id      createAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body schedjobdto.CreateSchedJobReq true "request data"
// @Success 201 {object} schedjobdto.CreateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs [post]
func (h *Handler) CreateAppSchedJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// UpdateAppSchedJob Updates a sched-job
// @Summary Updates a sched-job
// @Description Updates a sched-job
// @Tags    app_settings
// @Produce json
// @Id      updateAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs/{itemID} [put]
func (h *Handler) UpdateAppSchedJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// UpdateAppSchedJobStatus Updates sched-job status
// @Summary Updates sched-job status
// @Description Updates sched-job status
// @Tags    app_settings
// @Produce json
// @Id      updateAppSchedJobStatus
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobStatusReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs/{itemID}/status [put]
func (h *Handler) UpdateAppSchedJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// DeleteAppSchedJob Deletes sched-job
// @Summary Deletes sched-job
// @Description Deletes sched-job
// @Tags    app_settings
// @Produce json
// @Id      deleteAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.DeleteSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs/{itemID} [delete]
func (h *Handler) DeleteAppSchedJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeApp)
}

// ExecuteAppSchedJob Executes a sched job
// @Summary Executes a sched job
// @Description Executes a sched job
// @Tags    app_settings
// @Produce json
// @Id      executeAppSchedJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.ExecuteSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.ExecuteSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/sched-jobs/{itemID}/exec [post]
func (h *Handler) ExecuteAppSchedJob(ctx *gin.Context) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := schedjobdto.NewExecuteSchedJobReq()
	req.ID = jobID
	req.Scope = base.NewObjectScopeApp(appID, projectID)

	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SchedJobUC.ExecuteSchedJob(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
