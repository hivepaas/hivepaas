package filehandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler *authhandler.Handler
	fileUC      *fileuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	fileUC *fileuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		fileUC:      fileUC,
	}
}
