package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

// CreateProjectTag Creates a new project tag
// @Summary Creates a new project tag
// @Description Creates a new project tag
// @Tags    projects
// @Produce json
// @Id      createProjectTag
// @Param   projectID path string true "project ID"
// @Param   body body projectsettingsdto.CreateProjectTagReq true "request data"
// @Success 201 {object} projectsettingsdto.CreateProjectTagResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/tags [post]
func (h *Handler) CreateProjectTag(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewCreateProjectTagReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.CreateProjectTag(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteProjectTags Deletes project tags
// @Summary Deletes project tags
// @Description Deletes project tags
// @Tags    projects
// @Produce json
// @Id      deleteProjectTag
// @Param   projectID path string true "project ID"
// @Param   body body projectsettingsdto.DeleteProjectTagsReq true "request data"
// @Success 200 {object} projectsettingsdto.DeleteProjectTagsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/tags/delete [post]
func (h *Handler) DeleteProjectTags(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewDeleteProjectTagsReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.DeleteProjectTags(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
