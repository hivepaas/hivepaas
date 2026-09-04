package webhookhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc/webhookdto"
)

// webhookMaxBodySize matches the largest payload GitHub delivers, which is the
// most generous of the providers supported here.
const webhookMaxBodySize = 25 * 1024 * 1024

// HandleRepoWebhook Handles Repo webhook
// @Summary Handles Repo webhook
// @Description Handles Repo webhook
// @Tags    webhooks
// @Produce json
// @Id      handleRepoWebhook
// @Param   webhookID path string true "ID of repo-webhook or github-app"
// @Param   body body webhookdto.HandleRepoWebhookReq true "request data"
// @Success 200 {object} webhookdto.HandleRepoWebhookResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /webhooks/{webhookID} [post]
func (h *Handler) HandleRepoWebhook(ctx *gin.Context) {
	webhookID, err := h.ParseStringParam(ctx, "webhookID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// This route is unauthenticated: the body is only trusted once its signature
	// checks out, but it has to be read in full to check that. Cap it first, or
	// anyone who knows a webhook URL can make the process read an arbitrary body
	// into memory.
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, webhookMaxBodySize)

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
