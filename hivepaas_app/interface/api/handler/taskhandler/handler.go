package taskhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/taskuc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler *authhandler.Handler
	TaskUC      *taskuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	taskUC *taskuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		AuthHandler: authHandler,
		TaskUC:      taskUC,
	}
}
