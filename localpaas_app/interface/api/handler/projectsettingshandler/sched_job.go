package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ListSchedJob Lists sched-jobs
// @Summary Lists sched-jobs
// @Description Lists sched-jobs
// @Tags    project_settings
// @Produce json
// @Id      listProjectSchedJob
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} schedjobdto.ListSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs [get]
func (h *Handler) ListSchedJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}

// GetSchedJob Gets sched-job details
// @Summary Gets sched-job details
// @Description Gets sched-job details
// @Tags    project_settings
// @Produce json
// @Id      getProjectSchedJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.GetSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs/{itemID} [get]
func (h *Handler) GetSchedJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}

// CreateSchedJob Creates a new sched-job
// @Summary Creates a new sched-job
// @Description Creates a new sched-job
// @Tags    project_settings
// @Produce json
// @Id      createProjectSchedJob
// @Param   projectID path string true "project ID"
// @Param   body body schedjobdto.CreateSchedJobReq true "request data"
// @Success 201 {object} schedjobdto.CreateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs [post]
func (h *Handler) CreateSchedJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}

// UpdateSchedJob Updates sched-job
// @Summary Updates sched-job
// @Description Updates sched-job
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSchedJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs/{itemID} [put]
func (h *Handler) UpdateSchedJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}

// UpdateSchedJobStatus Updates sched-job status
// @Summary Updates sched-job status
// @Description Updates sched-job status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSchedJobStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body schedjobdto.UpdateSchedJobStatusReq true "request data"
// @Success 200 {object} schedjobdto.UpdateSchedJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs/{itemID}/status [put]
func (h *Handler) UpdateSchedJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}

// DeleteSchedJob Deletes sched-job
// @Summary Deletes sched-job
// @Description Deletes sched-job
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectSchedJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} schedjobdto.DeleteSchedJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/sched-jobs/{itemID} [delete]
func (h *Handler) DeleteSchedJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSchedJob, base.SettingScopeProject)
}
