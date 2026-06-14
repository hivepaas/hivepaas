package systemhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/taskuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	sysErrorUC  *syserroruc.UC
	taskUC      *taskuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	sysErrorUC *syserroruc.UC,
	taskUC *taskuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		sysErrorUC:  sysErrorUC,
		taskUC:      taskUC,
	}
}
