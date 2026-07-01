package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
)

// ListClusterVolume Lists cluster-volume settings
// @Summary Lists cluster-volume settings
// @Description Lists cluster-volume settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectClusterVolume
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} volumedto.ListVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes [get]
func (h *Handler) ListClusterVolume(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}

// GetClusterVolume Gets cluster-volume setting details
// @Summary Gets cluster-volume setting details
// @Description Gets cluster-volume setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectClusterVolume
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} volumedto.GetVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes/{itemID} [get]
func (h *Handler) GetClusterVolume(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}

// CreateClusterVolume Creates a new cluster-volume setting
// @Summary Creates a new cluster-volume setting
// @Description Creates a new cluster-volume setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectClusterVolume
// @Param   projectID path string true "project ID"
// @Param   body body volumedto.CreateVolumeReq true "request data"
// @Success 201 {object} volumedto.CreateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes [post]
func (h *Handler) CreateClusterVolume(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}

// UpdateClusterVolume Updates cluster-volume
// @Summary Updates cluster-volume
// @Description Updates cluster-volume
// @Tags    project_settings
// @Produce json
// @Id      updateProjectClusterVolume
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body volumedto.UpdateVolumeReq true "request data"
// @Success 200 {object} volumedto.UpdateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes/{itemID} [put]
func (h *Handler) UpdateClusterVolume(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}

// UpdateClusterVolumeStatus Updates cluster-volume status
// @Summary Updates cluster-volume status
// @Description Updates cluster-volume status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectClusterVolumeStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body volumedto.UpdateVolumeStatusReq true "request data"
// @Success 200 {object} volumedto.UpdateVolumeStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes/{itemID}/status [put]
func (h *Handler) UpdateClusterVolumeStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}

// DeleteClusterVolume Deletes cluster-volume setting
// @Summary Deletes cluster-volume setting
// @Description Deletes cluster-volume setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectClusterVolume
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} volumedto.DeleteVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/cluster-volumes/{itemID} [delete]
func (h *Handler) DeleteClusterVolume(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeProject)
}
