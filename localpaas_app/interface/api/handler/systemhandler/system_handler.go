package systemhandler

import (
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/syserroruc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc"
	"github.com/localpaas/localpaas/localpaas_app/usecase/taskuc"
)

type SystemHandler struct {
	*handler.BaseHandler
	authHandler *authhandler.AuthHandler
	sysErrorUC  *syserroruc.UC
	taskUC      *taskuc.UC
	lpAppUC     *lpappuc.UC
	traefikUC   *traefikuc.UC
}

func NewSystemHandler(
	baseHandler *handler.BaseHandler,
	authHandler *authhandler.AuthHandler,
	sysErrorUC *syserroruc.UC,
	taskUC *taskuc.UC,
	lpAppUC *lpappuc.UC,
	traefikUC *traefikuc.UC,
) *SystemHandler {
	return &SystemHandler{
		BaseHandler: baseHandler,
		authHandler: authHandler,
		sysErrorUC:  sysErrorUC,
		taskUC:      taskUC,
		lpAppUC:     lpAppUC,
		traefikUC:   traefikUC,
	}
}
