package userhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/useruc"
)

type UserHandler struct {
	*handler.BaseHandler
	authHandler *authhandler.AuthHandler
	userUC      *useruc.UC
}

func NewUserHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	userUC *useruc.UC,
) *UserHandler {
	return &UserHandler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		userUC:      userUC,
	}
}
