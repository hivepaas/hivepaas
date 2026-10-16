package systemhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/licenseuc/licensedto"
)

// GetLicense Gets the license in force
// @Summary Gets the license in force
// @Description The plan and status of the license in force, with the key's dates, how many of each limited
// @Description kind there are against the plan's limit (max null: no limit), and this installation's ID, which
// @Description a key must name. A key that is not in force, such as one for another installation, is described
// @Description too. Of the customer only the name is sent, never the email the key carries. With them are the
// @Description license portal's pages, for this installation, and whether the daily check for a newer key is on,
// @Description with how its last check went.
// @Tags    System
// @Produce json
// @Id      getLicense
// @Success 200 {object} licensedto.GetLicenseResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/license [get]
func (h *Handler) GetLicense(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := licensedto.NewGetLicenseReq()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.licenseUC.GetLicense(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateLicense Installs a license key
// @Summary Installs a license key
// @Description Verifies a license key and stores it in place of the one stored, whatever its plan, and answers
// @Description with the license it puts in force. Whitespace in the key is ignored. A key that does not verify,
// @Description is for another installation, or is past its grace is refused with 412, and the stored key is
// @Description kept. There is no way to remove a key.
// @Tags    System
// @Produce json
// @Id      updateLicense
// @Param   body body licensedto.UpdateLicenseReq true "request data"
// @Success 200 {object} licensedto.UpdateLicenseResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/license [put]
func (h *Handler) UpdateLicense(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := licensedto.NewUpdateLicenseReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.licenseUC.UpdateLicense(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RefreshLicense Asks the license portal for a newer key
// @Summary Asks the license portal for a newer key
// @Description Asks the license portal now whether there is a newer key than the one stored - a renewal, a plan
// @Description change, a paid key replacing a trial - and installs it, or learns that the key was moved to another
// @Description installation. Only the stored key and this installation's ID are sent. It asks whether or not the
// @Description daily check is switched off. Answers with what the portal said and the license in force after it,
// @Description as GET answers it: none, asking nothing, when no key is stored. A portal that cannot be reached is
// @Description 503, and one that refuses, or a newer key that cannot be installed, 412; either is recorded as the
// @Description last check's error.
// @Tags    System
// @Produce json
// @Id      refreshLicense
// @Success 200 {object} licensedto.RefreshLicenseResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Failure 503 {object} hperrors.ErrorInfo
// @Router  /system/license/refresh [post]
func (h *Handler) RefreshLicense(ctx *gin.Context) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.licenseUC.RefreshLicense(h.RequestCtx(ctx), auth, licensedto.NewRefreshLicenseReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
