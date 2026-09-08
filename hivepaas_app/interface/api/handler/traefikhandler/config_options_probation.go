package traefikhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
)

// ConfirmConfigOptions Confirms a Traefik config options change
// @Summary Confirms a Traefik config options change
// @Description Vouches for a config options change that is on trial, which stops it being undone
// @Tags    system_traefik
// @Produce json
// @Id      confirmTraefikConfigOptions
// @Param   body body traefiksettingsdto.ConfirmConfigOptionsReq true "request data"
// @Success 200 {object} traefiksettingsdto.ConfirmConfigOptionsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/traefik/config-options/confirm [post]
func (h *Handler) ConfirmConfigOptions(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Confirm config options").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefiksettingsdto.NewConfirmConfigOptionsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.ConfirmConfigOptions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RevertConfigOptions Reverts a Traefik config options change
// @Summary Reverts a Traefik config options change
// @Description Undoes a config options change that is on trial, without waiting for its deadline
// @Tags    system_traefik
// @Produce json
// @Id      revertTraefikConfigOptions
// @Param   body body traefiksettingsdto.RevertConfigOptionsReq true "request data"
// @Success 200 {object} traefiksettingsdto.RevertConfigOptionsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/traefik/config-options/revert [post]
func (h *Handler) RevertConfigOptions(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Revert config options").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := traefiksettingsdto.NewRevertConfigOptionsReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.traefikSettingsUC.RevertConfigOptions(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
