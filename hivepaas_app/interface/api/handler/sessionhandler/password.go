package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// LoginPasswordForgot Ask the system to send a password reset link via email
// @Summary Ask the system to send a password reset link via email
// @Description Ask the system to send a password reset link via email
// @Tags    sessions
// @Produce json
// @Id      loginPasswordForgot
// @Param   body body sessiondto.LoginPasswordForgotReq true "request data"
// @Success 200 {object} sessiondto.LoginPasswordForgotResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/login-password-forgot [post]
func (h *Handler) LoginPasswordForgot(ctx *gin.Context) {
	req := sessiondto.NewLoginPasswordForgotReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.LoginPasswordForgot(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
