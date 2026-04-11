package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
)

// ListDockerVolume Lists docker volumes
// @Summary Lists docker volumes
// @Description Lists docker volumes
// @Tags    project_settings
// @Produce json
// @Id      listProjectDockerVolume
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} volumedto.ListVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-volumes [get]
func (h *Handler) ListDockerVolume(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewListVolumeReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerVolumeUC.ListVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDockerVolume Gets docker volume details
// @Summary Gets docker volume details
// @Description Gets docker volume details
// @Tags    project_settings
// @Produce json
// @Id      getProjectDockerVolume
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-volumes/{volumeID} [get]
func (h *Handler) GetDockerVolume(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	volumeID, err := h.ParseStringParam(ctx, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewGetVolumeReq()
	req.VolumeID = volumeID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerVolumeUC.GetVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDockerVolumeInspection Gets docker volume details
// @Summary Gets docker volume details
// @Description Gets docker volume details
// @Tags    project_settings
// @Produce json
// @Id      getProjectVolumeInspection
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-volumes/{volumeID}/inspect [get]
func (h *Handler) GetDockerVolumeInspection(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	volumeID, err := h.ParseStringParam(ctx, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewGetVolumeInspectionReq()
	req.VolumeID = volumeID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerVolumeUC.GetVolumeInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateDockerVolume Creates a new docker volume
// @Summary Creates a new docker volume
// @Description Creates a new docker volume
// @Tags    project_settings
// @Produce json
// @Id      createProjectDockerVolume
// @Param   projectID path string true "project ID"
// @Param   body body volumedto.CreateVolumeReq true "request data"
// @Success 201 {object} volumedto.CreateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-volumes [post]
func (h *Handler) CreateDockerVolume(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewCreateVolumeReq()
	req.ProjectID = projectID
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerVolumeUC.CreateVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteDockerVolume Deletes a docker volume
// @Summary Deletes a docker volume
// @Description Deletes a docker volume
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectDockerVolume
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.DeleteVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-volumes/{volumeID} [delete]
func (h *Handler) DeleteDockerVolume(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	volumeID, err := h.ParseStringParam(ctx, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewDeleteVolumeReq()
	req.VolumeID = volumeID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerVolumeUC.DeleteVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
