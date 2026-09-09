package appbasehandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/auditloghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/basesettinghandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/taskhandler"
)

type Handler struct {
	*basesettinghandler.Handler
	AuditLogHandler *auditloghandler.Handler
	TaskHandler     *taskhandler.Handler
}

func New(
	baseSettingHandler *basesettinghandler.Handler,
	auditLogHandler *auditloghandler.Handler,
	taskHandler *taskhandler.Handler,
) *Handler {
	return &Handler{
		Handler:         baseSettingHandler,
		AuditLogHandler: auditLogHandler,
		TaskHandler:     taskHandler,
	}
}
