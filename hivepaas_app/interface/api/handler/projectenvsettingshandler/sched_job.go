package projectenvsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ListSchedJob Lists an env's scheduled jobs with its apps'
// @Summary Lists an env's scheduled jobs with its apps'
// @Description The env's own scheduled jobs - job sequences - and those of every app in the env, each with
// @Description its scope and, for an app's, the app.
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   jobType query string false "`jobType=type1,type2`"
// @Param   appId query string false "`appId=<an app of the env>`"
// @Param   status query string false "`status=active,disabled`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} schedjobdto.ListEnvSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs [get]
func (h *Handler) ListSchedJob(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := schedjobdto.NewListEnvSchedJobReq()
	req.Scope = &entity.ObjectScope{
		ScopeType: base.ObjectScopeProjectEnv, ProjectID: projectID, ProjectEnvID: projectEnvID,
	}
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SchedJobUC.ListEnvSchedJob(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetSchedJob Gets an env's scheduled job
// @Summary Gets an env's scheduled job
// @Description Gets an env's scheduled job
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.GetSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs/{itemID} [get]
func (h *Handler) GetSchedJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeProjectEnv)
}

// CreateSchedJob Creates an env's scheduled job
// @Summary Creates an env's scheduled job
// @Description Creates an env's scheduled job: a job sequence of its apps' jobs.
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body schedjobdto.CreateSchedJobReq true "request data"
// @Success 201 {object} schedjobdto.CreateSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs [post]
func (h *Handler) CreateSchedJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeProjectEnv)
}

// UpdateSchedJob Updates an env's scheduled job
// @Summary Updates an env's scheduled job
// @Description Updates an env's scheduled job
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs/{itemID} [put]
func (h *Handler) UpdateSchedJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeProjectEnv)
}

// UpdateSchedJobStatus Enables or disables an env's scheduled job
// @Summary Enables or disables an env's scheduled job
// @Description Enables or disables an env's scheduled job
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSchedJobStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobStatusReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs/{itemID}/status [put]
func (h *Handler) UpdateSchedJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSchedJob, base.ObjectScopeProjectEnv)
}

// DeleteSchedJob Deletes an env's scheduled job
// @Summary Deletes an env's scheduled job
// @Description Deletes an env's scheduled job
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.DeleteSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs/{itemID} [delete]
func (h *Handler) DeleteSchedJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSchedJob, base.ObjectScopeProjectEnv)
}

// ExecuteSchedJob Runs an env's scheduled job now
// @Summary Runs an env's scheduled job now
// @Description Runs an env's scheduled job now, outside its schedule.
// @Tags    project_env_settings
// @Produce json
// @Id      executeProjectEnvSchedJob
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.ExecuteSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.ExecuteSchedJobResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/sched-jobs/{itemID}/exec [post]
func (h *Handler) ExecuteSchedJob(ctx *gin.Context) {
	auth, projectID, projectEnvID, jobID, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeExecute, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := schedjobdto.NewExecuteSchedJobReq()
	req.ID = jobID
	req.Scope = &entity.ObjectScope{
		ScopeType: base.ObjectScopeProjectEnv, ProjectID: projectID, ProjectEnvID: projectEnvID,
	}
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
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
