package filehandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc"
)

type FileHandler struct {
	*handler.BaseHandler
	authHandler *authhandler.AuthHandler
	fileUC      *fileuc.UC
}

func NewFileHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	fileUC *fileuc.UC,
) *FileHandler {
	return &FileHandler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		fileUC:      fileUC,
	}
}
