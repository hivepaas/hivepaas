package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

// ListAppCronJob Lists cron-jobs
// @Summary Lists cron-jobs
// @Description Lists cron-jobs
// @Tags    app_settings
// @Produce json
// @Id      listAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cronjobdto.ListCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs [get]
func (h *Handler) ListAppCronJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// GetAppCronJob Gets cron-job details
// @Summary Gets cron-job details
// @Description Gets cron-job details
// @Tags    app_settings
// @Produce json
// @Id      getAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cronjobdto.GetCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID} [get]
func (h *Handler) GetAppCronJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// CreateAppCronJob Creates a new cron-job
// @Summary Creates a new cron-job
// @Description Creates a new cron-job
// @Tags    app_settings
// @Produce json
// @Id      createAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body cronjobdto.CreateCronJobReq true "request data"
// @Success 201 {object} cronjobdto.CreateCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs [post]
func (h *Handler) CreateAppCronJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// UpdateAppCronJob Updates a cron-job
// @Summary Updates a cron-job
// @Description Updates a cron-job
// @Tags    app_settings
// @Produce json
// @Id      updateAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cronjobdto.UpdateCronJobReq true "request data"
// @Success 200 {object} cronjobdto.UpdateCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID} [put]
func (h *Handler) UpdateAppCronJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// UpdateAppCronJobStatus Updates cron-job status
// @Summary Updates cron-job status
// @Description Updates cron-job status
// @Tags    app_settings
// @Produce json
// @Id      updateAppCronJobStatus
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cronjobdto.UpdateCronJobStatusReq true "request data"
// @Success 200 {object} cronjobdto.UpdateCronJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID}/status [put]
func (h *Handler) UpdateAppCronJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// DeleteAppCronJob Deletes cron-job
// @Summary Deletes cron-job
// @Description Deletes cron-job
// @Tags    app_settings
// @Produce json
// @Id      deleteAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cronjobdto.DeleteCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID} [delete]
func (h *Handler) DeleteAppCronJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeApp)
}

// ExecuteAppCronJob Executes a cron job
// @Description Executes a cron job
// @Tags    app_settings
// @Produce json
// @Id      executeAppCronJob
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cronjobdto.ExecuteCronJobReq true "request data"
// @Success 200 {object} cronjobdto.ExecuteCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID}/exec [post]
func (h *Handler) ExecuteAppCronJob(ctx *gin.Context) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := cronjobdto.NewExecuteCronJobReq()
	req.ID = jobID
	req.Scope = base.NewSettingScopeApp(appID, projectID)

	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.CronJobUC.ExecuteCronJob(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
