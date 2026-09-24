package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// GetAppDockerAPISettings Gets what an app may do through the Docker API
// @Summary Gets what an app may do through the Docker API
// @Description Gets the app's Docker API access: the images its children may run, the directories they
// @Description share, the networks they join, the groups of endpoints allowed, and the limits.
// @Tags    app_settings
// @Produce json
// @Id      getAppDockerAPISettings
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appsettingsdto.GetAppDockerAPISettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/docker-api-settings [get]
func (h *Handler) GetAppDockerAPISettings(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewGetAppDockerAPISettingsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.GetAppDockerAPISettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateAppDockerAPISettings Updates what an app may do through the Docker API
// @Summary Updates what an app may do through the Docker API
// @Description Updates the app's Docker API access. Turning it on, or letting it do more, needs Write
// @Description permission on the Cluster module; narrowing it or turning it off needs Write on the app.
// @Tags    app_settings
// @Produce json
// @Id      updateAppDockerAPISettings
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appsettingsdto.UpdateAppDockerAPISettingsReq true "request data"
// @Success 200 {object} appsettingsdto.UpdateAppDockerAPISettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/docker-api-settings [put]
func (h *Handler) UpdateAppDockerAPISettings(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewUpdateAppDockerAPISettingsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.UpdateAppDockerAPISettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
