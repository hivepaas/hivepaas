package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// DevModeLogin Login user for dev mode
// @Summary Login user for dev mode
// @Description Login user for dev mode
// @Tags    sessions
// @Produce json
// @Id      devModeLogin
// @Param   userId query string false "user ID to login"
// @Success 200 {object} sessiondto.DevModeLoginResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /internal/auth/dev-mode-login [post]
func (h *Handler) DevModeLogin(ctx *gin.Context) {
	req := sessiondto.NewDevModeLoginReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.DevModeLogin(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
