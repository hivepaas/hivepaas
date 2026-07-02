package webhookhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc/webhookdto"
)

// HandleRepoWebhook Handles Repo webhook
// @Summary Handles Repo webhook
// @Description Handles Repo webhook
// @Tags    webhooks
// @Produce json
// @Id      handleRepoWebhook
// @Param   webhookID path string true "ID of repo-webhook or github-app"
// @Param   body body webhookdto.HandleRepoWebhookReq true "request data"
// @Success 200 {object} webhookdto.HandleRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /webhooks/{webhookID} [post]
func (h *Handler) HandleRepoWebhook(ctx *gin.Context) {
	webhookID, err := h.ParseStringParam(ctx, "webhookID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := webhookdto.NewHandleRepoWebhookReq()
	req.Request = ctx.Request
	req.ID = webhookID

	resp, err := h.webhookUC.HandleRepoWebhook(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
