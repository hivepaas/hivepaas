package hivepaashandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

// GetProject Gets the project HivePaaS runs in
// @Summary Gets the project HivePaaS runs in
// @Description The project of HivePaaS's own apps and those it provisions, which the projects list leaves out.
// @Description Admin only.
// @Tags    HivePaaS
// @Produce json
// @Id      getHivePaaSProject
// @Success 200 {object} hpappdto.GetHpProjectResp
// @Failure 403 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/hivepaas/project [get]
func (h *Handler) GetProject(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	if auth.User.Role != base.UserRoleAdmin {
		h.RenderError(ctx, hperrors.NewForbidden("Get the HivePaaS project").
			WithMsgLog("only admin can perform this action"))
		return
	}

	req := hpappdto.NewGetHpProjectReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.hpAppUC.GetHpProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
