package projectenvhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/projectbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc"
)

type Handler struct {
	*projectbasehandler.Handler
	projectEnvUC *projectenvuc.UC
}

func New(
	baseHandler *projectbasehandler.Handler,
	projectEnvUC *projectenvuc.UC,
) *Handler {
	return &Handler{
		Handler:      baseHandler,
		projectEnvUC: projectEnvUC,
	}
}
