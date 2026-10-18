package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// RefreshManagedCredentials Gives a managed app its archive storage's current credentials
// @Summary Gives a managed app its archive storage's current credentials
// @Description The credentials were copied into the app when it was made: after rotating the storage's key, this
// @Description gives the database the new one. The secrets that changed are mounted again, restarting its task.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedCredentialsRefresh
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.RefreshManagedCredentialsReq true "request data"
// @Success 200 {object} appmanageddto.RefreshManagedCredentialsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/refresh-credentials [post]
func (h *Handler) RefreshManagedCredentials(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewRefreshManagedCredentialsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.RefreshManagedCredentials(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
