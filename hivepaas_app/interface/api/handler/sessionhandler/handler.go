package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	sessionUC   *sessionuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	sessionUC *sessionuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		sessionUC:   sessionUC,
	}
}

// GetMe Gets session info of the current user
// @Summary Gets session info of the current user
// @Description Gets session info of the current user
// @Tags    sessions
// @Produce json
// @Id      getMe
// @Param   getAccesses query string false "`getAccesses=true/false`"
// @Success 200 {object} sessiondto.GetMeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /sessions/me [get]
func (h *Handler) GetMe(ctx *gin.Context) {
	user, err := h.authHandler.GetCurrentUser(ctx)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sessiondto.NewGetMeReq()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.GetMe(h.RequestCtx(ctx), user, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
