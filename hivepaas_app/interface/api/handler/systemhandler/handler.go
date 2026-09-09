package systemhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/auditloghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/taskhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/syserroruc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/sysstatusuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler     *authhandler.Handler
	auditLogHandler *auditloghandler.Handler
	taskHandler     *taskhandler.Handler
	sysErrorUC      *syserroruc.UC
	sysStatusUC     *sysstatusuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	auditLogHandler *auditloghandler.Handler,
	taskHandler *taskhandler.Handler,
	sysErrorUC *syserroruc.UC,
	sysStatusUC *sysstatusuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:     baseHandler,
		authHandler:     authHandler,
		auditLogHandler: auditLogHandler,
		taskHandler:     taskHandler,
		sysErrorUC:      sysErrorUC,
		sysStatusUC:     sysStatusUC,
	}
}
