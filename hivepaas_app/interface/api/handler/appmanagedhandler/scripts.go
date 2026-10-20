package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// GetManagedScripts Gets what a managed database's template offers its engine's scripts
// @Summary Gets what a managed database's template offers its engine's scripts
// @Description What the installation's revision of the database's template offers of what its engine owns -
// @Description its scripts and artifacts - against what the database runs: up-to-date, update-available,
// @Description incompatible or unavailable, the changes, and the members out of step. A cluster is read on its
// @Description app; a member is refused.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedScriptsGet
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appmanageddto.GetManagedScriptsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/scripts [get]
func (h *Handler) GetManagedScripts(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewGetManagedScriptsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.GetManagedScripts(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpgradeManagedScripts Upgrades a managed database's engine scripts
// @Summary Upgrades a managed database's engine scripts
// @Description Replaces the scripts and artifacts of the database, its members and its verifier with what the
// @Description installation's revision of its template offers, in one transaction, then deploys each, the member
// @Description that writes last: each restarts once. The fingerprint is the one reviewed (GET .../managed/scripts);
// @Description another is refused. A member, a database still restoring, and scripts with nothing to apply are
// @Description refused.
// @Tags    Managed apps
// @Accept  json
// @Produce json
// @Id      appManagedScriptsUpgrade
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.UpgradeManagedScriptsReq true "request data"
// @Success 200 {object} appmanageddto.UpgradeManagedScriptsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/scripts/upgrade [post]
func (h *Handler) UpgradeManagedScripts(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewUpgradeManagedScriptsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.UpgradeManagedScripts(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
