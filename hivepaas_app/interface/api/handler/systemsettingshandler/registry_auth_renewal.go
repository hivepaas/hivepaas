package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc/registryauthrenewaldto"
)

// GetRegistryAuthRenewalSettings Gets registry auth renewal settings
// @Summary Gets registry auth renewal settings
// @Description Gets registry auth renewal settings
// @Tags    System settings
// @Produce json
// @Id      getRegistryAuthRenewalSettings
// @Success 200 {object} registryauthrenewaldto.GetRegistryAuthRenewalResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry-auth-renewal [get]
func (h *Handler) GetRegistryAuthRenewalSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeRegistryAuthRenewal,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registryauthrenewaldto.NewGetRegistryAuthRenewalReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryAuthRenewalUC.GetRegistryAuthRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateRegistryAuthRenewalSettings Updates registry auth renewal settings
// @Summary Updates registry auth renewal settings
// @Description Updates registry auth renewal settings
// @Tags    System settings
// @Produce json
// @Id      updateRegistryAuthRenewalSettings
// @Param   body body registryauthrenewaldto.UpdateRegistryAuthRenewalReq true "request data"
// @Success 200 {object} registryauthrenewaldto.UpdateRegistryAuthRenewalResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry-auth-renewal [put]
func (h *Handler) UpdateRegistryAuthRenewalSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeRegistryAuthRenewal,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registryauthrenewaldto.NewUpdateRegistryAuthRenewalReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryAuthRenewalUC.UpdateRegistryAuthRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteRegistryAuthRenewal Runs the registry auth renewal now
// @Summary Runs the registry auth renewal now
// @Description Runs the registry auth renewal now
// @Tags    System settings
// @Produce json
// @Id      executeRegistryAuthRenewal
// @Param   body body registryauthrenewaldto.ExecuteRegistryAuthRenewalReq true "request data"
// @Success 200 {object} registryauthrenewaldto.ExecuteRegistryAuthRenewalResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry-auth-renewal/exec [post]
func (h *Handler) ExecuteRegistryAuthRenewal(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeExecute},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeRegistryAuthRenewal,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registryauthrenewaldto.NewExecuteRegistryAuthRenewalReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryAuthRenewalUC.ExecuteRegistryAuthRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
