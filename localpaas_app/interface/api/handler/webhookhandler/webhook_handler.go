package webhookhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/webhookuc"
)

type WebhookHandler struct {
	*handler.BaseHandler
	webhookUC *webhookuc.UC
}

func NewWebhookHandler(
	baseHandler *handler.BaseHandler,
	webhookUC *webhookuc.UC,
) *WebhookHandler {
	return &WebhookHandler{
		BaseHandler: baseHandler,
		webhookUC:   webhookUC,
	}
}
