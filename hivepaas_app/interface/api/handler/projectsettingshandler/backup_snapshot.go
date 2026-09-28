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
// @Tags    Project settings
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
// @Tags    Project settings
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
// @Tags    Project settings
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

// ListBackupSnapshotEntries Lists what a backup snapshot holds
// @Summary Lists what a backup snapshot holds
// @Description Lists a directory of a backup snapshot the scope sees: its files and directories, directories first
// @Tags    project_settings
// @Produce json
// @Id      listProjectBackupSnapshotEntries
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "snapshot record ID"
// @Param   path query string false "a directory inside the snapshot; empty for its root"
// @Success 200 {object} backupsnapshotdto.ListBackupSnapshotEntriesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots/{itemID}/entries [get]
func (h *Handler) ListBackupSnapshotEntries(ctx *gin.Context) {
	h.Handler.ListBackupSnapshotEntries(ctx, base.ObjectScopeProject)
}

// RestoreBackupSnapshot Restores a backup snapshot into an app
// @Summary Restores a backup snapshot into an app
// @Description Records a task that puts a snapshot the scope sees back into an app the caller may change: a command
// @Description snapshot through a command run in the app, a volume snapshot into a volume the app mounts
// @Tags    project_settings
// @Accept  json
// @Produce json
// @Id      restoreProjectBackupSnapshot
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "snapshot record ID"
// @Param   body body backupsnapshotdto.RestoreBackupSnapshotReq true "request data"
// @Success 200 {object} backupsnapshotdto.RestoreBackupSnapshotResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots/{itemID}/restore [post]
func (h *Handler) RestoreBackupSnapshot(ctx *gin.Context) {
	h.Handler.RestoreBackupSnapshot(ctx, base.ObjectScopeProject)
}

// DownloadBackupSnapshotFile Downloads a file of a backup snapshot
// @Summary Downloads a file of a backup snapshot
// @Description Streams a file of a snapshot the scope sees, to a caller who may write on the snapshot's owner;
// @Description the download is recorded in the audit log
// @Tags    project_settings
// @Produce octet-stream
// @Id      downloadProjectBackupSnapshotFile
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "snapshot record ID"
// @Param   path query string true "a file inside the snapshot"
// @Success 200 {file} binary
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/backup-snapshots/{itemID}/download [get]
func (h *Handler) DownloadBackupSnapshotFile(ctx *gin.Context) {
	h.Handler.DownloadBackupSnapshotFile(ctx, base.ObjectScopeProject)
}
