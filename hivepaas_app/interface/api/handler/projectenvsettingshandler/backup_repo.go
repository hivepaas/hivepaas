package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

// ListBackupRepo Lists backup repo settings
// @Summary Lists backup repo settings
// @Description Lists backup repo settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} backuprepodto.ListBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos [get]
func (h *Handler) ListBackupRepo(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// GetBackupRepo Gets backup repo setting details
// @Summary Gets backup repo setting details
// @Description Gets backup repo setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.GetBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID} [get]
func (h *Handler) GetBackupRepo(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// CreateBackupRepo Creates a new backup repo setting
// @Summary Creates a new backup repo setting
// @Description Creates a new backup repo setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body backuprepodto.CreateBackupRepoReq true "request data"
// @Success 201 {object} backuprepodto.CreateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos [post]
func (h *Handler) CreateBackupRepo(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// UpdateBackupRepo Updates backup repo
// @Summary Updates backup repo
// @Description Updates backup repo
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID} [put]
func (h *Handler) UpdateBackupRepo(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// ChangeBackupRepoPassword Changes a backup repository's password
// @Summary Changes a backup repository's password
// @Description Re-encrypts the backup repository with a new password. The repository itself is
// @Description re-encrypted first, so the previous password stops working as soon as this succeeds.
// @Tags    project_env_settings
// @Produce json
// @Id      changeProjectEnvBackupRepoPassword
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.ChangeRepoPasswordReq true "request data"
// @Success 200 {object} backuprepodto.ChangeRepoPasswordResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID}/password [put]
func (h *Handler) ChangeBackupRepoPassword(ctx *gin.Context) {
	h.BackupRepoChangePassword(ctx, base.ObjectScopeProjectEnv)
}

// UpdateBackupRepoStatus Updates backup repo status
// @Summary Updates backup repo status
// @Description Updates backup repo status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvBackupRepoStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoStatusReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID}/status [put]
func (h *Handler) UpdateBackupRepoStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// DeleteBackupRepo Deletes backup repo setting
// @Summary Deletes backup repo setting
// @Description Deletes backup repo setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.DeleteBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID} [delete]
func (h *Handler) DeleteBackupRepo(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeProjectEnv)
}

// CleanupBackupRepo Cleans up a backup repository
// @Summary Cleans up a backup repository
// @Description Applies the repository's retention policy, removing the snapshots it expires, then
// @Description reconciles the stored snapshot records against what the repository still holds.
// @Tags    project_env_settings
// @Produce json
// @Id      cleanupProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.CleanupBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID}/cleanup [post]
func (h *Handler) CleanupBackupRepo(ctx *gin.Context) {
	h.BackupRepoCleanup(ctx, base.ObjectScopeProjectEnv)
}

// SyncBackupRepo Syncs a backup repository back into its setting
// @Summary Syncs a backup repository back into its setting
// @Description Reads the repository and adopts what it finds: the options it is configured with,
// @Description and the snapshots it holds. Use it after the repository was changed outside the
// @Description app. Nothing in the repository is modified.
// @Tags    project_env_settings
// @Produce json
// @Id      syncProjectEnvBackupRepo
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.SyncBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/backup-repos/{itemID}/sync [post]
func (h *Handler) SyncBackupRepo(ctx *gin.Context) {
	h.BackupRepoSync(ctx, base.ObjectScopeProjectEnv)
}
