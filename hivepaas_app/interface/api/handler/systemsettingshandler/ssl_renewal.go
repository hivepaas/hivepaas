package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

// GetSSLRenewalSettings Gets SSL renewal settings
// @Summary Gets SSL renewal settings
// @Description Gets SSL renewal settings
// @Tags    system_settings
// @Produce json
// @Id      getSSLRenewalSettings
// @Success 200 {object} sslrenewaldto.GetSSLRenewalResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/ssl-renewal [get]
func (h *Handler) GetSSLRenewalSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSSLRenewal,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslrenewaldto.NewGetSSLRenewalReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSLRenewalUC.GetSSLRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateSSLRenewalSettings Updates SSL renewal settings
// @Summary Updates SSL renewal settings
// @Description Updates SSL renewal settings
// @Tags    system_settings
// @Produce json
// @Id      updateSSLRenewalSettings
// @Param   body body sslrenewaldto.UpdateSSLRenewalReq true "request data"
// @Success 200 {object} sslrenewaldto.UpdateSSLRenewalResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/ssl-renewal [put]
func (h *Handler) UpdateSSLRenewalSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSSLRenewal,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslrenewaldto.NewUpdateSSLRenewalReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSLRenewalUC.UpdateSSLRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteSSLRenewal Executes the renewal
// @Summary Executes the renewal
// @Description Executes the renewal
// @Tags    system_settings
// @Produce json
// @Id      executeSSLRenewal
// @Param   body body sslrenewaldto.ExecuteSSLRenewalReq true "request data"
// @Success 200 {object} sslrenewaldto.ExecuteSSLRenewalResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/ssl-renewal/exec [post]
func (h *Handler) ExecuteSSLRenewal(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSSLRenewal,
		Action:         base.ActionTypeExecute,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sslrenewaldto.NewExecuteSSLRenewalReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSLRenewalUC.ExecuteSSLRenewal(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
