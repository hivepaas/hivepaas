package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/getstarteduc/getstarteddto"
)

// GetDashboardCert Tells where the dashboard's certificate stands
// @Summary Tells where the dashboard's certificate stands
// @Description For the Get started card, while nextStep is hivepaas/get-started: todo, obtaining, failed
// @Description (with the error) or done. The first answer that finds it done also clears the step, for
// @Description every admin.
// @Tags    system
// @Produce json
// @Id      getDashboardCert
// @Success 200 {object} getstarteddto.GetDashboardCertResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/get-started/dashboard-cert [get]
func (h *Handler) GetDashboardCert(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.getStartedUC.GetDashboardCert(h.RequestCtx(ctx), auth, getstarteddto.NewGetDashboardCertReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RequestDashboardCert Asks for the dashboard's certificate now
// @Summary Asks for the dashboard's certificate now
// @Description For the Get started card: asks for a certificate for the dashboard's domain, past
// @Description the wait a failed attempt leaves. Refused with ERR_CONFLICT while one is being
// @Description obtained. Answers with where the certificate stands.
// @Tags    system
// @Produce json
// @Id      requestDashboardCert
// @Success 200 {object} getstarteddto.RequestDashboardCertResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 409 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/get-started/dashboard-cert [post]
func (h *Handler) RequestDashboardCert(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.getStartedUC.RequestDashboardCert(h.RequestCtx(ctx), auth,
		getstarteddto.NewRequestDashboardCertReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DismissGetStarted Closes the Get started card
// @Summary Closes the Get started card
// @Description Clears the installation step, which hides the card for every admin.
// @Tags    system
// @Produce json
// @Id      dismissGetStarted
// @Success 200 {object} getstarteddto.DismissResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/get-started/dismiss [post]
func (h *Handler) DismissGetStarted(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.getStartedUC.Dismiss(h.RequestCtx(ctx), auth, getstarteddto.NewDismissReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
