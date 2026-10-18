package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// RestoreManagedApp Restores a managed database into a new app
// @Summary Restores a managed database into a new app
// @Description To the latest, a time, a restore point or a base backup, inside its restore window. The new app is
// @Description made from the same template, version and parameters, on the data volume given, and archives to its
// @Description own prefix; the source is only read. The task returned sees the restore through.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedRestore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.RestoreManagedReq true "request data"
// @Success 200 {object} appmanageddto.RestoreManagedResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/restore [post]
func (h *Handler) RestoreManagedApp(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	// A restore makes an app in the env, as creating one does.
	if _, _, _, _, err = h.GetAuthInEnv(ctx, base.ActionTypeWrite, false); err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewRestoreManagedReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.RestoreManagedApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
