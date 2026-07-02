package supporthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/supportuc/supportdto"
)

// CreateFeedback Creates a feedback
// @Summary Creates a feedback
// @Description Creates a feedback
// @Tags    support_feedbacks
// @Produce json
// @Id      createFeedback
// @Param   body body supportdto.CreateFeedbackReq true "request data"
// @Success 200 {object} supportdto.CreateFeedbackResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /support/feedbacks [post]
func (h *Handler) CreateFeedback(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := supportdto.NewCreateFeedbackReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.supportUC.CreateFeedback(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
