package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// BuildCommandTemplate Builds a command template to get command string
// @Summary Builds a command template to get command string
// @Description Builds a command template to get command string
// @Tags    app_settings
// @Produce json
// @Id      buildAppCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "command template ID"
// @Success 200 {object} appsettingsdto.BuildCommandTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/command-templates/{itemID}/build [post]
func (h *Handler) BuildCommandTemplate(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeExecute, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appsettingsdto.NewBuildCommandTemplateReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	req.CommandID = itemID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appSettingsUC.BuildCommandTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
