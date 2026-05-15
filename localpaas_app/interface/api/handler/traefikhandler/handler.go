package traefikhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefiksettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler       *authhandler.Handler
	traefikUC         *traefikuc.UC
	traefikSettingsUC *traefiksettingsuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	traefikUC *traefikuc.UC,
	traefikSettingsUC *traefiksettingsuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:       baseHandler,
		authHandler:       authHandler,
		traefikUC:         traefikUC,
		traefikSettingsUC: traefikSettingsUC,
	}
}
