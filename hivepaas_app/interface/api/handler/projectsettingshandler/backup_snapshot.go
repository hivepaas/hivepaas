package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// ListBackupSnapshot Lists the backup snapshots the scope sees
// @Summary Lists the backup snapshots the scope sees
// @Description Lists the snapshots of the backup repositories the scope sees, those of its apps and repositories
// @Description the caller may read, newest first
// @Tags    project_settings
// @Produce json
// @Id      listProjectBackupSnapshot
// @Param   projectID path string true "project ID"
// @Param   repo query []string false "repository IDs" collectionFormat(multi)
// @Param   app query []string false "app IDs" collectionFormat(multi)
// @Param   tag query []string false "tags key:value, all of which must match" collectionFormat(multi)
// @Param   fromDate query string false "from date, YYYY-MM-DD"
// @Param   toDate query string false "to date, YYYY-MM-DD, included"
// @Param   search query string false "short ID or description"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Success 200 {object} backupsnapshotdto.ListBackupSnapshotResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots [get]
func (h *Handler) ListBackupSnapshot(ctx *gin.Context) {
	h.Handler.ListBackupSnapshot(ctx, base.ObjectScopeProject)
}

// GetBackupSnapshot Gets a backup snapshot
// @Summary Gets a backup snapshot
// @Description Gets a backup snapshot the scope sees
// @Tags    project_settings
// @Produce json
// @Id      getProjectBackupSnapshot
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "snapshot record ID"
// @Success 200 {object} backupsnapshotdto.GetBackupSnapshotResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots/{itemID} [get]
func (h *Handler) GetBackupSnapshot(ctx *gin.Context) {
	h.Handler.GetBackupSnapshot(ctx, base.ObjectScopeProject)
}

// DeleteBackupSnapshot Deletes a backup snapshot
// @Summary Deletes a backup snapshot
// @Description Deletes a snapshot from its repository, then its record; one already gone counts as deleted
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectBackupSnapshot
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "snapshot record ID"
// @Success 200 {object} backupsnapshotdto.DeleteBackupSnapshotResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots/{itemID} [delete]
func (h *Handler) DeleteBackupSnapshot(ctx *gin.Context) {
	h.Handler.DeleteBackupSnapshot(ctx, base.ObjectScopeProject)
}
