package hivepaashandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/logginguc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler     *authhandler.Handler
	hpAppUC         *hpappuc.UC
	hpAppSettingsUC *hpappsettingsuc.UC
	loggingUC       *logginguc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	hpAppUC *hpappuc.UC,
	hpAppSettingsUC *hpappsettingsuc.UC,
	loggingUC *logginguc.UC,
) *Handler {
	return &Handler{
		BaseHandler:     baseHandler,
		authHandler:     authHandler,
		hpAppUC:         hpAppUC,
		hpAppSettingsUC: hpAppSettingsUC,
		loggingUC:       loggingUC,
	}
}
