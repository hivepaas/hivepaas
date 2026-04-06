package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
)

// ListNetwork Lists networks
// @Summary Lists networks
// @Description Lists networks
// @Tags    project_settings
// @Produce json
// @Id      listProjectNetwork
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} networkdto.ListNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/networks [get]
func (h *ProjectHandler) ListNetwork(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.networkUC.ListNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNetwork Gets network details
// @Summary Gets network details
// @Description Gets network details
// @Tags    project_settings
// @Produce json
// @Id      getProjectNetwork
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/networks/{networkID} [get]
func (h *ProjectHandler) GetNetwork(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.networkUC.GetNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNetworkInspection Gets network details
// @Summary Gets network details
// @Description Gets network details
// @Tags    project_settings
// @Produce json
// @Id      getProjectNetworkInspection
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/networks/{networkID}/inspect [get]
func (h *ProjectHandler) GetNetworkInspection(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeRead, true)
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

	resp, err := h.networkUC.GetNetworkInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateNetwork Creates a new network setting
// @Summary Creates a new network setting
// @Description Creates a new network setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectNetwork
// @Param   projectID path string true "project ID"
// @Param   body body networkdto.CreateNetworkReq true "request data"
// @Success 201 {object} networkdto.CreateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/networks [post]
func (h *ProjectHandler) CreateNetwork(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeWrite, true)
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

	resp, err := h.networkUC.CreateNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteNetwork Deletes a network
// @Summary Deletes a network
// @Description Deletes a network
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectNetwork
// @Param   projectID path string true "project ID"
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.DeleteNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/networks/{networkID} [delete]
func (h *ProjectHandler) DeleteNetwork(ctx *gin.Context) {
	auth, projectID, err := h.getAuth(ctx, base.ActionTypeWrite, true)
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

	resp, err := h.networkUC.DeleteNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
