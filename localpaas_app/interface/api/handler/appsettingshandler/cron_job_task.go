package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

// ListAppCronJobTask Lists cron-job tasks
// @Summary Lists cron-job tasks
// @Description Lists cron-job tasks
// @Tags    app_settings
// @Produce json
// @Id      listAppCronJobTask
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cronjobdto.ListCronJobTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID}/tasks [get]
func (h *Handler) ListAppCronJobTask(ctx *gin.Context) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := cronjobdto.NewListCronJobTaskReq()
	req.JobID = jobID
	base.NewSettingScopeApp(appID, projectID)
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.CronJobUC.ListCronJobTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppCronJobTask Gets a cron-job task
// @Summary Gets a cron-job task
// @Description Gets a cron-job task
// @Tags    app_settings
// @Produce json
// @Id      getAppCronJobTask
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   taskID path string true "task ID"
// @Success 200 {object} cronjobdto.GetCronJobTaskResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID}/tasks/{taskID} [get]
func (h *Handler) GetAppCronJobTask(ctx *gin.Context) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := cronjobdto.NewGetCronJobTaskReq()
	req.TaskID = taskID
	req.JobID = jobID
	base.NewSettingScopeApp(appID, projectID)
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.CronJobUC.GetCronJobTask(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppCronJobTaskLogs Gets logs of a cron-job task
// @Summary Gets logs of a cron-job task
// @Description Gets logs of a cron-job task
// @Tags    app_settings
// @Produce json
// @Id      getAppCronJobTaskLogs
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   taskID path string true "task ID"
// @Success 200 {object} cronjobdto.GetCronJobTaskLogsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/cron-jobs/{itemID}/tasks/{taskID}/logs [get]
func (h *Handler) GetAppCronJobTaskLogs(ctx *gin.Context, mel *melody.Melody) {
	auth, projectID, appID, jobID, err := h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	taskID, err := h.ParseStringParam(ctx, "taskID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := cronjobdto.NewGetCronJobTaskLogsReq()
	req.TaskID = taskID
	req.JobID = jobID
	req.Scope = base.NewSettingScopeApp(appID, projectID)
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	isWebsocketReq := h.IsWebsocketRequest(ctx)
	if !isWebsocketReq {
		req.Follow = false // Not a websocket request, we don't support `follow` flag
	}

	resp, err := h.CronJobUC.GetCronJobTaskLogs(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if !isWebsocketReq {
		// Not a websocket request, return data via body
		ctx.JSON(http.StatusOK, resp)
	} else {
		h.StreamAppLogs(ctx, resp.Data.Logs, resp.Data.LogChan, resp.Data.LogChanCloser, mel)
	}
}
