package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/mcpuc/mcpdto"
)

// GetMCPSettings Gets MCP server settings
// @Summary Gets MCP server settings
// @Description Gets whether HivePaaS serves the Model Context Protocol at <API base path>/mcp.
// @Tags    system_settings
// @Produce json
// @Id      getMCPSettings
// @Success 200 {object} mcpdto.GetMCPSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/mcp [get]
func (h *Handler) GetMCPSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeMCP,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := mcpdto.NewGetMCPSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.MCPUC.GetMCPSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateMCPSettings Updates MCP server settings
// @Summary Updates MCP server settings
// @Description Turns the MCP server on or off.
// @Tags    system_settings
// @Produce json
// @Id      updateMCPSettings
// @Param   body body mcpdto.UpdateMCPSettingsReq true "request data"
// @Success 200 {object} mcpdto.UpdateMCPSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/mcp [put]
func (h *Handler) UpdateMCPSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeMCP,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := mcpdto.NewUpdateMCPSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.MCPUC.UpdateMCPSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
