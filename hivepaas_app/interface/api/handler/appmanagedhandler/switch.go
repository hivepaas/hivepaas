package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// GetManagedSwitch Lists what switching to a restored database would repoint
// @Summary Lists what switching to a restored database would repoint
// @Description The database this one was restored from, and the apps of the environment that use it: by a
// @Description variable naming its key, by its id among their previews' databases, or as their deployments'
// @Description restore point. Only for a restored database, once it is ready.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedSwitchGet
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appmanageddto.GetManagedSwitchResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/switch [get]
func (h *Handler) GetManagedSwitch(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewGetManagedSwitchReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.GetManagedSwitch(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// SwitchManaged Switches consumers to a restored database
// @Summary Switches consumers to a restored database
// @Description The consumers chosen name this database instead of the one it was restored from: their
// @Description variables, their previews' databases and their deployments' restore point. Their variables are
// @Description applied to their running services; with redeploy, each whose variables changed is deployed too.
// @Description The source keeps running, its jobs paused.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedSwitch
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.SwitchManagedReq true "request data"
// @Success 200 {object} appmanageddto.SwitchManagedResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/switch [post]
func (h *Handler) SwitchManaged(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewSwitchManagedReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.SwitchManaged(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
