package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// CreateRestorePoint Names a moment of a managed database, to restore to exactly
// @Summary Names a moment of a managed database, to restore to exactly
// @Description The point is made, and listed once the WAL it is in is archived; it can then be restored to for
// @Description as long as retention keeps the base backup before it, or until keepUntil, which holds retention
// @Description back. Names are labels: letters, digits, '.', '_' and '-', at most 63, not starting with "hp-".
// @Tags    Managed apps
// @Produce json
// @Id      appManagedRestorePointCreate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.CreateRestorePointReq true "request data"
// @Success 200 {object} appmanageddto.CreateRestorePointResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/restore-points [post]
func (h *Handler) CreateRestorePoint(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewCreateRestorePointReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.CreateRestorePoint(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateRestorePoint Changes how long a restore point is kept, and its note
// @Summary Changes how long a restore point is kept, and its note
// @Description Both as given: a null keepUntil keeps the point no longer, and retention moves on at its next run.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedRestorePointUpdate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   name path string true "restore point name"
// @Param   body body appmanageddto.UpdateRestorePointReq true "request data"
// @Success 200 {object} appmanageddto.UpdateRestorePointResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/restore-points/{name} [put]
func (h *Handler) UpdateRestorePoint(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "name")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewUpdateRestorePointReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	req.Name = name
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.UpdateRestorePoint(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteRestorePoint Removes a restore point's record
// @Summary Removes a restore point's record
// @Description The archive is not touched, and the mark stays in the WAL, harmlessly: restores target an LSN.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedRestorePointDelete
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   name path string true "restore point name"
// @Success 200 {object} appmanageddto.DeleteRestorePointResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/restore-points/{name} [delete]
func (h *Handler) DeleteRestorePoint(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "name")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewDeleteRestorePointReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	req.Name = name
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.DeleteRestorePoint(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
