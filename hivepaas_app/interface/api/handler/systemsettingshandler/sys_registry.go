package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryuc/registrydto"
)

// GetRegistrySettings Gets registry settings
// @Summary Gets registry settings
// @Description Gets registry settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemRegistrySettings
// @Success 200 {object} registrydto.GetRegistrySettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry [get]
func (h *Handler) GetRegistrySettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeRegistry,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registrydto.NewGetRegistrySettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryUC.GetRegistrySettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateRegistrySettings Updates registry settings
// @Summary Updates registry settings
// @Description Updates registry settings
// @Tags    system_settings
// @Produce json
// @Id      updateSystemRegistrySettings
// @Param   body body registrydto.UpdateRegistrySettingsReq true "request data"
// @Success 200 {object} registrydto.UpdateRegistrySettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry [put]
func (h *Handler) UpdateRegistrySettings(ctx *gin.Context) {
	auth, err := h.registryWriteAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registrydto.NewUpdateRegistrySettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryUC.UpdateRegistrySettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ProbeRegistryDomain Checks what answers at the registry's address
// @Summary Checks what answers at the registry's address
// @Description Reports whether a proxy answers instead of the registry, which is advisory
// @Tags    system_settings
// @Produce json
// @Id      probeSystemRegistryDomain
// @Param   body body registrydto.ProbeDomainReq true "request data"
// @Success 200 {object} registrydto.ProbeDomainResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry/probe-domain [post]
func (h *Handler) ProbeRegistryDomain(ctx *gin.Context) {
	auth, err := h.registryWriteAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := &registrydto.ProbeDomainReq{}
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryUC.ProbeRegistryDomain(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CheckRegistryPush Uploads a large blob to the registry and throws it away
// @Summary Uploads a large blob to the registry and throws it away
// @Description Reports whether anything in front of the registry limits request bodies
// @Tags    system_settings
// @Produce json
// @Id      checkSystemRegistryPush
// @Param   body body registrydto.PushCheckReq true "request data"
// @Success 200 {object} registrydto.PushCheckResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry/push-check [post]
func (h *Handler) CheckRegistryPush(ctx *gin.Context) {
	auth, err := h.registryWriteAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := &registrydto.PushCheckReq{}
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryUC.CheckRegistryPush(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RotateRegistryCredential Issues a new registry password
// @Summary Issues a new registry password
// @Description The previous password keeps working until the grace period ends
// @Tags    system_settings
// @Produce json
// @Id      rotateSystemRegistryCredential
// @Success 200 {object} registrydto.RotateCredentialResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/registry/rotate-credential [post]
func (h *Handler) RotateRegistryCredential(ctx *gin.Context) {
	auth, err := h.registryWriteAuth(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryUC.RotateRegistryCredential(h.RequestCtx(ctx), auth)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// registryWriteAuth is the check every registry action shares: each of them
// changes the registry, or spends its bandwidth, so each needs the same right as
// saving the settings.
func (h *Handler) registryWriteAuth(ctx *gin.Context) (*basedto.Auth, error) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeRegistry,
	})
	return auth, hperrors.Wrap(err)
}
