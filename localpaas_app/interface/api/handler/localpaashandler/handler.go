package localpaashandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappsettingsuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc"
)

type Handler struct {
	*handler.BaseHandler
	authHandler     *authhandler.Handler
	lpAppUC         *lpappuc.UC
	lpAppSettingsUC *lpappsettingsuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.Handler,
	lpAppUC *lpappuc.UC,
	lpAppSettingsUC *lpappsettingsuc.UC,
) *Handler {
	return &Handler{
		BaseHandler:     baseHandler,
		authHandler:     authHandler,
		lpAppUC:         lpAppUC,
		lpAppSettingsUC: lpAppSettingsUC,
	}
}
