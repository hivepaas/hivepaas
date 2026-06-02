package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accessiblebyprojectsuc/accessiblebyprojectsdto"
)

// GetAccessibleByProjects Gets accessible by projects of a setting
// @Summary Gets accessible by projects of a setting
// @Description Gets accessible by projects of a setting
// @Tags    settings
// @Produce json
// @Id      getAccessibleByProjects
// @Param   itemID path string true "setting ID"
// @Success 200 {object} accessiblebyprojectsdto.GetAccessibleByProjectsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/{itemID}/accessible-by-projects [get]
func (h *Handler) GetAccessibleByProjects(ctx *gin.Context) {
	settingID, err := h.ParseStringParam(ctx, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSettings,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := accessiblebyprojectsdto.NewGetAccessibleByProjectsReq()
	req.SettingID = settingID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.AccessibleByProjectsUC.GetAccessibleByProjects(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateAccessibleByProjects Updates accessible by projects of a setting
// @Summary Updates accessible by projects of a setting
// @Description Updates accessible by projects of a setting
// @Tags    settings
// @Produce json
// @Id      updateAccessibleByProjects
// @Param   itemID path string true "setting ID"
// @Param   body body accessiblebyprojectsdto.UpdateAccessibleByProjectsReq true "request data"
// @Success 200 {object} accessiblebyprojectsdto.UpdateAccessibleByProjectsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/{itemID}/accessible-by-projects [put]
func (h *Handler) UpdateAccessibleByProjects(ctx *gin.Context) {
	settingID, err := h.ParseStringParam(ctx, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSettings,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := accessiblebyprojectsdto.NewUpdateAccessibleByProjectsReq()
	req.SettingID = settingID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.AccessibleByProjectsUC.UpdateAccessibleByProjects(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
