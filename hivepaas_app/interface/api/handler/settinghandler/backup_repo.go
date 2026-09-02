package settinghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ListBackupRepo Lists backup repositories
// @Summary Lists backup repositories
// @Description Lists backup repositories
// @Tags    settings
// @Produce json
// @Id      listSettingBackupRepo
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} backuprepodto.ListBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos [get]
func (h *Handler) ListBackupRepo(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}

// GetBackupRepo Gets backup repository details
// @Summary Gets backup repository details
// @Description Gets backup repository details
// @Tags    settings
// @Produce json
// @Id      getSettingBackupRepo
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.GetBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos/{itemID} [get]
func (h *Handler) GetBackupRepo(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}

// CreateBackupRepo Creates a new backup repository
// @Summary Creates a new backup repository
// @Description Creates a new backup repository on the storage backend, or adopts an existing one
// @Description when `importExisting` is set, in which case the snapshots already in the repository
// @Description are read back and stored.
// @Tags    settings
// @Produce json
// @Id      createSettingBackupRepo
// @Param   body body backuprepodto.CreateBackupRepoReq true "request data"
// @Success 201 {object} backuprepodto.CreateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos [post]
func (h *Handler) CreateBackupRepo(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}

// UpdateBackupRepo Updates a backup repository
// @Summary Updates a backup repository
// @Description Updates a backup repository
// @Tags    settings
// @Produce json
// @Id      updateSettingBackupRepo
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos/{itemID} [put]
func (h *Handler) UpdateBackupRepo(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}

// ChangeBackupRepoPassword Changes a backup repository's password
// @Summary Changes a backup repository's password
// @Description Re-encrypts the backup repository with a new password. The repository itself is
// @Description re-encrypted first, so the previous password stops working as soon as this succeeds.
// @Tags    settings
// @Produce json
// @Id      changeSettingBackupRepoPassword
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.ChangeRepoPasswordReq true "request data"
// @Success 200 {object} backuprepodto.ChangeRepoPasswordResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos/{itemID}/password [put]
func (h *Handler) ChangeBackupRepoPassword(ctx *gin.Context) {
	h.BackupRepoChangePassword(ctx, base.ObjectScopeGlobal)
}

// UpdateBackupRepoStatus Updates a backup repository's status
// @Summary Updates a backup repository's status
// @Description Updates a backup repository's status
// @Tags    settings
// @Produce json
// @Id      updateSettingBackupRepoStatus
// @Param   itemID path string true "setting ID"
// @Param   body body backuprepodto.UpdateBackupRepoStatusReq true "request data"
// @Success 200 {object} backuprepodto.UpdateBackupRepoStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos/{itemID}/status [put]
func (h *Handler) UpdateBackupRepoStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}

// DeleteBackupRepo Deletes a backup repository
// @Summary Deletes a backup repository
// @Description Deletes the backup repository setting. The data on the storage backend is kept.
// @Tags    settings
// @Produce json
// @Id      deleteSettingBackupRepo
// @Param   itemID path string true "setting ID"
// @Success 200 {object} backuprepodto.DeleteBackupRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /settings/backup-repos/{itemID} [delete]
func (h *Handler) DeleteBackupRepo(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeBackupRepo, base.ObjectScopeGlobal)
}
