package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

// GetProjectUserAccesses Gets user accesses on the project
// @Summary Gets user accesses on the project
// @Description Gets user accesses on the project
// @Tags    project_settings
// @Produce json
// @Id      getProjectUserAccesses
// @Param   projectID path string true "project ID"
// @Success 200 {object} projectsettingsdto.GetUserAccessesResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/user-accesses [get]
func (h *Handler) GetProjectUserAccesses(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewGetUserAccessesReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.GetUserAccesses(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateProjectUserAccesses Updates user accesses on the project
// @Summary Updates user accesses on the project
// @Description Updates user accesses on the project
// @Tags    project_settings
// @Produce json
// @Id      updateProjectUserAccesses
// @Param   projectID path string true "project ID"
// @Param   body body projectsettingsdto.UpdateUserAccessesReq true "request data"
// @Success 201 {object} projectsettingsdto.UpdateUserAccessesResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/user-accesses [put]
func (h *Handler) UpdateProjectUserAccesses(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewUpdateUserAccessesReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.UpdateUserAccesses(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
