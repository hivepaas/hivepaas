package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

// ListBackupRepo Lists backup repo settings
// @Summary Lists backup repo settings
// @Description Lists backup repo settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectBackupRepo
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} backuprepodto.ListBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos [get]
func (h *Handler) ListBackupRepo(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}

// GetBackupRepo Gets backup repo setting details
// @Summary Gets backup repo setting details
// @Description Gets backup repo setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectBackupRepo
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.GetBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos/{itemID} [get]
func (h *Handler) GetBackupRepo(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}

// CreateBackupRepo Creates a new backup repo setting
// @Summary Creates a new backup repo setting
// @Description Creates a new backup repo setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectBackupRepo
// @Param   projectID path string true "project ID"
// @Param   body body backuprepodto.CreateBackupRepoReq true "request data"
// @Success 201 {object} backuprepodto.CreateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos [post]
func (h *Handler) CreateBackupRepo(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}

// UpdateBackupRepo Updates backup repo
// @Summary Updates backup repo
// @Description Updates backup repo
// @Tags    project_settings
// @Produce json
// @Id      updateProjectBackupRepo
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos/{itemID} [put]
func (h *Handler) UpdateBackupRepo(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}

// UpdateBackupRepoStatus Updates backup repo status
// @Summary Updates backup repo status
// @Description Updates backup repo status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectBackupRepoStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoStatusReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos/{itemID}/status [put]
func (h *Handler) UpdateBackupRepoStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}

// DeleteBackupRepo Deletes backup repo setting
// @Summary Deletes backup repo setting
// @Description Deletes backup repo setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectBackupRepo
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.DeleteBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-repos/{itemID} [delete]
func (h *Handler) DeleteBackupRepo(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProject)
}
