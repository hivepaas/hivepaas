package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
)

// ListClusterNetwork Lists cluster-network settings
// @Summary Lists cluster-network settings
// @Description Lists cluster-network settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectClusterNetwork
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} networkdto.ListNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks [get]
func (h *Handler) ListClusterNetwork(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}

// GetClusterNetwork Gets cluster-network setting details
// @Summary Gets cluster-network setting details
// @Description Gets cluster-network setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectClusterNetwork
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} networkdto.GetNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks/{itemID} [get]
func (h *Handler) GetClusterNetwork(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}

// CreateClusterNetwork Creates a new cluster-network setting
// @Summary Creates a new cluster-network setting
// @Description Creates a new cluster-network setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectClusterNetwork
// @Param   projectID path string true "project ID"
// @Param   body body networkdto.CreateNetworkReq true "request data"
// @Success 201 {object} networkdto.CreateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks [post]
func (h *Handler) CreateClusterNetwork(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}

// UpdateClusterNetwork Updates cluster-network
// @Summary Updates cluster-network
// @Description Updates cluster-network
// @Tags    project_settings
// @Produce json
// @Id      updateProjectClusterNetwork
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body networkdto.UpdateNetworkReq true "request data"
// @Success 200 {object} networkdto.UpdateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks/{itemID} [put]
func (h *Handler) UpdateClusterNetwork(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}

// UpdateClusterNetworkStatus Updates cluster-network status
// @Summary Updates cluster-network status
// @Description Updates cluster-network status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectClusterNetworkStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body networkdto.UpdateNetworkStatusReq true "request data"
// @Success 200 {object} networkdto.UpdateNetworkStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks/{itemID}/status [put]
func (h *Handler) UpdateClusterNetworkStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}

// DeleteClusterNetwork Deletes cluster-network setting
// @Summary Deletes cluster-network setting
// @Description Deletes cluster-network setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectClusterNetwork
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} networkdto.DeleteNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-networks/{itemID} [delete]
func (h *Handler) DeleteClusterNetwork(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeProject)
}
