package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ListSchedJob Lists sched-jobs
// @Summary Lists sched-jobs
// @Description Lists sched-jobs
// @Tags    settings
// @Produce json
// @Id      listSettingSchedJob
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} schedjobdto.ListSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs [get]
func (h *Handler) ListSchedJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// GetSchedJob Gets sched-job details
// @Summary Gets sched-job details
// @Description Gets sched-job details
// @Tags    settings
// @Produce json
// @Id      getSettingSchedJob
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.GetSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs/{itemID} [get]
func (h *Handler) GetSchedJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// CreateSchedJob Creates a new sched-job
// @Summary Creates a new sched-job
// @Description Creates a new sched-job
// @Tags    settings
// @Produce json
// @Id      createSettingSchedJob
// @Param   body body schedjobdto.CreateSchedJobReq true "request data"
// @Success 201 {object} schedjobdto.CreateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs [post]
func (h *Handler) CreateSchedJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// UpdateSchedJob Updates sched-job
// @Summary Updates sched-job
// @Description Updates sched-job
// @Tags    settings
// @Produce json
// @Id      updateSettingSchedJob
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs/{itemID} [put]
func (h *Handler) UpdateSchedJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// UpdateSchedJobStatus Updates sched-job status
// @Summary Updates sched-job status
// @Description Updates sched-job status
// @Tags    settings
// @Produce json
// @Id      updateSettingSchedJobStatus
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobStatusReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs/{itemID}/status [put]
func (h *Handler) UpdateSchedJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// DeleteSchedJob Deletes sched-job
// @Summary Deletes sched-job
// @Description Deletes sched-job
// @Tags    settings
// @Produce json
// @Id      deleteSettingSchedJob
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.DeleteSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs/{itemID} [delete]
func (h *Handler) DeleteSchedJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeGlobal)
}

// SchedJobCalcNextRuns Calculates next runs of the job
// @Summary Calculates next runs of the job
// @Description Calculates next runs of the job
// @Tags    settings
// @Produce json
// @Id      schedJobCalcNextRuns
// @Param   body body schedjobdto.CalcNextRunsReq true "request data"
// @Success 200 {object} schedjobdto.CalcNextRunsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/sched-jobs/calc-next-runs [post]
func (h *Handler) SchedJobCalcNextRuns(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := schedjobdto.NewCalcNextRunsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SchedJobUC.CalcNextRuns(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
