package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
)

// ListVolume Lists volumes
// @Summary Lists volumes
// @Description Lists volumes
// @Tags    project_settings
// @Produce json
// @Id      listProjectVolume
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} volumedto.ListVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/volumes [get]
func (h *ProjectHandler) ListVolume(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.volumeUC.ListVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetVolume Gets volume details
// @Summary Gets volume details
// @Description Gets volume details
// @Tags    project_settings
// @Produce json
// @Id      getProjectVolume
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/volumes/{volumeID} [get]
func (h *ProjectHandler) GetVolume(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.volumeUC.GetVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetVolumeInspection Gets volume details
// @Summary Gets volume details
// @Description Gets volume details
// @Tags    project_settings
// @Produce json
// @Id      getProjectVolumeInspection
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/volumes/{volumeID}/inspect [get]
func (h *ProjectHandler) GetVolumeInspection(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.volumeUC.GetVolumeInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateVolume Creates a new volume
// @Summary Creates a new volume
// @Description Creates a new volume
// @Tags    project_settings
// @Produce json
// @Id      createProjectVolume
// @Param   projectID path string true "project ID"
// @Param   body body volumedto.CreateVolumeReq true "request data"
// @Success 201 {object} volumedto.CreateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/volumes [post]
func (h *ProjectHandler) CreateVolume(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeWrite, true)
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

	resp, err := h.volumeUC.CreateVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteVolume Deletes a volume
// @Summary Deletes a volume
// @Description Deletes a volume
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectVolume
// @Param   projectID path string true "project ID"
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.DeleteVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/volumes/{volumeID} [delete]
func (h *ProjectHandler) DeleteVolume(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeWrite, true)
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

	resp, err := h.volumeUC.DeleteVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
