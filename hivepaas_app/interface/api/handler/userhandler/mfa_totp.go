package userhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

// BeginMFATotpSetup Begins MFA TOTP authenticator setup
// @Summary Begins MFA TOTP authenticator setup
// @Description Begins MFA TOTP authenticator setup
// @Tags    users
// @Produce json
// @Id      beginMFATotpSetup
// @Param   body body userdto.BeginMFATotpSetupReq true "request data"
// @Success 200 {object} userdto.BeginMFATotpSetupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/mfa/totp-begin-setup [post]
func (h *Handler) BeginMFATotpSetup(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil && !errors.Is(err, apperrors.ErrUserNotCompleteMFASetup) {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewBeginMFATotpSetupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.BeginMFATotpSetup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CompleteMFATotpSetup Completes MFA TOTP authenticator setup
// @Summary Completes MFA TOTP authenticator setup
// @Description Completes MFA TOTP authenticator setup
// @Tags    users
// @Produce json
// @Id      completeMFATotpSetup
// @Param   body body userdto.CompleteMFATotpSetupReq true "request data"
// @Success 200 {object} userdto.CompleteMFATotpSetupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/mfa/totp-complete-setup [post]
func (h *Handler) CompleteMFATotpSetup(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil && !errors.Is(err, apperrors.ErrUserNotCompleteMFASetup) {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewCompleteMFATotpSetupReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.CompleteMFATotpSetup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RemoveMFATotp Removes MFA TOTP authenticator setup
// @Summary Removes MFA TOTP authenticator setup
// @Description Removes MFA TOTP authenticator setup
// @Tags    users
// @Produce json
// @Id      removeMFATotp
// @Param   body body userdto.RemoveMFATotpReq true "request data"
// @Success 200 {object} userdto.RemoveMFATotpResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/mfa/totp-remove [post]
func (h *Handler) RemoveMFATotp(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := userdto.NewRemoveMFATotpReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.userUC.RemoveMFATotp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
