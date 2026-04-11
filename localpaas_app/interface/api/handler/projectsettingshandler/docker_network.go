package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
)

// ListDockerNetwork Lists docker networks
// @Summary Lists docker networks
// @Description Lists docker networks
// @Tags    project_settings
// @Produce json
// @Id      listProjectDockerNetwork
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} networkdto.ListNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-networks [get]
func (h *Handler) ListDockerNetwork(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewListNetworkReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerNetworkUC.ListNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDockerNetwork Gets docker network details
// @Summary Gets docker network details
// @Description Gets docker network details
// @Tags    project_settings
// @Produce json
// @Id      getProjectDockerNetwork
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-networks/{networkID} [get]
func (h *Handler) GetDockerNetwork(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	networkID, err := h.ParseStringParam(ctx, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewGetNetworkReq()
	req.NetworkID = networkID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerNetworkUC.GetNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDockerNetworkInspection Gets docker network details
// @Summary Gets docker network details
// @Description Gets docker network details
// @Tags    project_settings
// @Produce json
// @Id      getProjectDockerNetworkInspection
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-networks/{networkID}/inspect [get]
func (h *Handler) GetDockerNetworkInspection(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	networkID, err := h.ParseStringParam(ctx, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewGetNetworkInspectionReq()
	req.NetworkID = networkID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerNetworkUC.GetNetworkInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateDockerNetwork Creates a new docker network
// @Summary Creates a new docker network
// @Description Creates a new docker network
// @Tags    project_settings
// @Produce json
// @Id      createProjectDockerNetwork
// @Param   projectID path string true "project ID"
// @Param   body body networkdto.CreateNetworkReq true "request data"
// @Success 201 {object} networkdto.CreateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-networks [post]
func (h *Handler) CreateDockerNetwork(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewCreateNetworkReq()
	req.ProjectID = projectID
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerNetworkUC.CreateNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteDockerNetwork Deletes a docker network
// @Summary Deletes a docker network
// @Description Deletes a docker network
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectDockerNetwork
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.DeleteNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/docker-networks/{networkID} [delete]
func (h *Handler) DeleteDockerNetwork(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	networkID, err := h.ParseStringParam(ctx, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewDeleteNetworkReq()
	req.NetworkID = networkID
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.dockerNetworkUC.DeleteNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
