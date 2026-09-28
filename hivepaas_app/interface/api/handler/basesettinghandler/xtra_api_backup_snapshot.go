package basesettinghandler

import (
	"mime"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// authBackupSnapshotScope is the viewer of a scope's snapshots, acting with
// action, and the scope. Reading a scope's snapshots needs what reading its
// backup repositories does; which snapshots it then sees is the use case's.
func (h *Handler) authBackupSnapshotScope(
	ctx *gin.Context,
	scopeType base.ObjectScopeType,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, scope *entity.ObjectScope, itemID string, err error) {
	scope = &entity.ObjectScope{ScopeType: scopeType}
	switch scopeType { //nolint:exhaustive
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, action, paramName)
	case base.ObjectScopeProjectEnv:
		auth, scope.ProjectID, scope.ProjectEnvID, itemID, err = h.GetAuthProjectEnvSettings(ctx, action, paramName)
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.ProjectEnvID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx,
			action, paramName)
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeBackupRepo, action, paramName)
	default:
		err = hperrors.Wrap(hperrors.ErrObjectScopeInvalid).WithParam("Scope", scopeType)
	}
	return auth, scope, itemID, hperrors.Wrap(err)
}

func (h *Handler) ListBackupSnapshot(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, _, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewListBackupSnapshotReq()
	req.Scope = scope
	if err = h.ParseAndValidateRequest(ctx, req, req.PagingReq()); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.ListBackupSnapshot(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GetBackupSnapshot(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, itemID, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewGetBackupSnapshotReq()
	req.Scope = scope
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.GetBackupSnapshot(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteBackupSnapshot(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, itemID, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeDelete, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewDeleteBackupSnapshotReq()
	req.Scope = scope
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.DeleteBackupSnapshot(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) ListBackupSnapshotEntries(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, itemID, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewListBackupSnapshotEntriesReq()
	req.Scope = scope
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.ListBackupSnapshotEntries(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// RestoreBackupSnapshot reads the snapshot as the scope's viewer does; writing on
// the app it goes into is the use case's to check.
func (h *Handler) RestoreBackupSnapshot(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, itemID, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewRestoreBackupSnapshotReq()
	req.Scope = scope
	req.ID = itemID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.RestoreBackupSnapshot(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// DownloadBackupSnapshotFile writes a file of a snapshot to the response. It
// needs write on the scope, and on the snapshot's owner, which the use case
// checks. A failure once the headers are sent cuts the body short of its
// Content-Length, which a client takes for a failed download.
func (h *Handler) DownloadBackupSnapshotFile(ctx *gin.Context, scopeType base.ObjectScopeType) {
	auth, scope, itemID, err := h.authBackupSnapshotScope(ctx, scopeType, base.ActionTypeWrite, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req := backupsnapshotdto.NewDownloadBackupSnapshotFileReq()
	req.Scope = scope
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.BackupSnapshotUC.DownloadBackupSnapshotFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	writeDownloadHeaders(ctx, resp)
	ctx.Status(http.StatusOK)
	if err = resp.Write(h.RequestCtx(ctx), ctx.Writer); err != nil {
		_ = ctx.Error(err)
	}
}

// writeDownloadHeaders makes the response an attachment of the file's name and size.
func writeDownloadHeaders(ctx *gin.Context, resp *backupsnapshotdto.DownloadBackupSnapshotFileResp) {
	ctx.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": resp.FileName}))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", strconv.FormatInt(resp.SizeBytes, 10))
}
