package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

// ConfirmRoutingSettings Confirms a routing change that is on trial
// @Summary Confirms a routing change that is on trial
// @Description Keeps a routing change that would otherwise be reverted at its deadline
// @Tags    system_hivepaas
// @Produce json
// @Id      confirmHivePaaSRoutingSettings
// @Param   body body hpappsettingsdto.ConfirmRoutingSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.ConfirmRoutingSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/routing-settings/confirm [post]
func (h *Handler) ConfirmRoutingSettings(ctx *gin.Context) {
	auth, err := h.authorizeRoutingProbation(ctx, "Confirm routing settings")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := hpappsettingsdto.NewConfirmRoutingSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.ConfirmRoutingSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RevertRoutingSettings Reverts a routing change that is on trial
// @Summary Reverts a routing change that is on trial
// @Description Undoes an unconfirmed routing change now, without waiting for its deadline
// @Tags    system_hivepaas
// @Produce json
// @Id      revertHivePaaSRoutingSettings
// @Param   body body hpappsettingsdto.RevertRoutingSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.RevertRoutingSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/routing-settings/revert [post]
func (h *Handler) RevertRoutingSettings(ctx *gin.Context) {
	auth, err := h.authorizeRoutingProbation(ctx, "Revert routing settings")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := hpappsettingsdto.NewRevertRoutingSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.RevertRoutingSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// authorizeRoutingProbation gates every end of confirm-or-revert the way the
// update itself is gated: whoever may change these settings is who may vouch for
// a change or take it back.
func (h *Handler) authorizeRoutingProbation(ctx *gin.Context, action string) (*basedto.Auth, error) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if auth.User.Role != base.UserRoleAdmin {
		return nil, hperrors.NewForbidden(action).WithMsgLog("only admin can perform this action")
	}
	return auth, nil
}

// ConfirmServiceSettings Confirms a service settings change that is on trial
// @Summary Confirms a service settings change that is on trial
// @Description Keeps a proxy settings change that would otherwise be reverted at its deadline
// @Tags    system_hivepaas
// @Produce json
// @Id      confirmHivePaaSServiceSettings
// @Param   body body hpappsettingsdto.ConfirmServiceSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.ConfirmServiceSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/service-settings/confirm [post]
func (h *Handler) ConfirmServiceSettings(ctx *gin.Context) {
	auth, err := h.authorizeRoutingProbation(ctx, "Confirm service settings")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := hpappsettingsdto.NewConfirmServiceSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.ConfirmServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RevertServiceSettings Reverts a service settings change that is on trial
// @Summary Reverts a service settings change that is on trial
// @Description Undoes an unconfirmed proxy settings change now, without waiting for its deadline
// @Tags    system_hivepaas
// @Produce json
// @Id      revertHivePaaSServiceSettings
// @Param   body body hpappsettingsdto.RevertServiceSettingsReq true "request data"
// @Success 200 {object} hpappsettingsdto.RevertServiceSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/service-settings/revert [post]
func (h *Handler) RevertServiceSettings(ctx *gin.Context) {
	auth, err := h.authorizeRoutingProbation(ctx, "Revert service settings")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := hpappsettingsdto.NewRevertServiceSettingsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppSettingsUC.RevertServiceSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
