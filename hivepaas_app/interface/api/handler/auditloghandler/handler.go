package auditloghandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/auditloguc"
)

type Handler struct {
	*handler.BaseHandler
	AuthHandler *authhandler.Handler
	AuditLogUC  *auditloguc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	auditLogUC *auditloguc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		AuthHandler: authHandler,
		AuditLogUC:  auditLogUC,
	}
}
