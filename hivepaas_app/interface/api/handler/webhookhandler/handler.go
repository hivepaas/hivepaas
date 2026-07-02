package webhookhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/webhookuc"
)

type Handler struct {
	*handler.BaseHandler
	webhookUC *webhookuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	webhookUC *webhookuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		webhookUC:   webhookUC,
	}
}
