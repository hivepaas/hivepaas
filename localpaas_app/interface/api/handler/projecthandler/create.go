package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

// CreateProject Creates a new project
// @Summary Creates a new project
// @Description Creates a new project
// @Tags    projects
// @Produce json
// @Id      createProject
// @Param   body body projectdto.CreateProjectReq true "request data"
// @Success 201 {object} projectdto.CreateProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects [post]
func (h *Handler) CreateProject(ctx *gin.Context) {
	auth, _, err := h.GetAuth(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewCreateProjectReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.CreateProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// UpdateProject Updates a project
// @Summary Updates a project
// @Description Updates a project
// @Tags    projects
// @Produce json
// @Id      updateProject
// @Param   projectID path string true "project ID"
// @Param   body body projectdto.UpdateProjectReq true "request data"
// @Success 200 {object} projectdto.UpdateProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID} [put]
func (h *Handler) UpdateProject(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewUpdateProjectReq()
	req.ID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.UpdateProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateProjectPhoto Updates project photo
// @Summary Updates project photo
// @Description Updates project photo
// @Tags    projects
// @Produce json
// @Id      updateProjectPhoto
// @Param   projectID path string true "project ID"
// @Param   body body projectdto.UpdateProjectPhotoReq true "request data"
// @Success 200 {object} projectdto.UpdateProjectPhotoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/photo [put]
func (h *Handler) UpdateProjectPhoto(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewUpdateProjectPhotoReq()
	req.ID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.UpdateProjectPhoto(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteProject Deletes a project
// @Summary Deletes a project
// @Description Deletes a project
// @Tags    projects
// @Produce json
// @Id      deleteProject
// @Param   projectID path string true "project ID"
// @Success 200 {object} projectdto.DeleteProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID} [delete]
func (h *Handler) DeleteProject(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeDelete, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewDeleteProjectReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.DeleteProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
