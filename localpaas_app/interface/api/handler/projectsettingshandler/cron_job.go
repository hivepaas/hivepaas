package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
)

// ListCronJob Lists cron-jobs
// @Summary Lists cron-jobs
// @Description Lists cron-jobs
// @Tags    projects
// @Produce json
// @Id      listProjectCronJob
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cronjobdto.ListCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs [get]
func (h *Handler) ListCronJob(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}

// GetCronJob Gets cron-job details
// @Summary Gets cron-job details
// @Description Gets cron-job details
// @Tags    projects
// @Produce json
// @Id      getProjectCronJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cronjobdto.GetCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs/{itemID} [get]
func (h *Handler) GetCronJob(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}

// CreateCronJob Creates a new cron-job
// @Summary Creates a new cron-job
// @Description Creates a new cron-job
// @Tags    projects
// @Produce json
// @Id      createProjectCronJob
// @Param   projectID path string true "project ID"
// @Param   body body cronjobdto.CreateCronJobReq true "request data"
// @Success 201 {object} cronjobdto.CreateCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs [post]
func (h *Handler) CreateCronJob(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}

// UpdateCronJob Updates cron-job
// @Summary Updates cron-job
// @Description Updates cron-job
// @Tags    projects
// @Produce json
// @Id      updateProjectCronJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cronjobdto.UpdateCronJobReq true "request data"
// @Success 200 {object} cronjobdto.UpdateCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs/{itemID} [put]
func (h *Handler) UpdateCronJob(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}

// UpdateCronJobStatus Updates cron-job status
// @Summary Updates cron-job status
// @Description Updates cron-job status
// @Tags    projects
// @Produce json
// @Id      updateProjectCronJobStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body cronjobdto.UpdateCronJobStatusReq true "request data"
// @Success 200 {object} cronjobdto.UpdateCronJobStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs/{itemID}/status [put]
func (h *Handler) UpdateCronJobStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}

// DeleteCronJob Deletes cron-job
// @Summary Deletes cron-job
// @Description Deletes cron-job
// @Tags    projects
// @Produce json
// @Id      deleteProjectCronJob
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cronjobdto.DeleteCronJobResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cron-jobs/{itemID} [delete]
func (h *Handler) DeleteCronJob(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCronJob, base.SettingScopeProject)
}
