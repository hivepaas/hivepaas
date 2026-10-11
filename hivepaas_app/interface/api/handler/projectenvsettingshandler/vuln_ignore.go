package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/vulnignoreuc/vulnignoredto"
)

// ListVulnIgnore Lists vulnerability ignores
// @Summary Lists vulnerability ignores
// @Description Lists vulnerability ignores
// @Tags    Project env settings
// @Produce json
// @Id      listProjectEnvVulnIgnore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} vulnignoredto.ListVulnIgnoreResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores [get]
func (h *Handler) ListVulnIgnore(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}

// GetVulnIgnore Gets vulnerability ignore details
// @Summary Gets vulnerability ignore details
// @Description Gets vulnerability ignore details
// @Tags    Project env settings
// @Produce json
// @Id      getProjectEnvVulnIgnore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} vulnignoredto.GetVulnIgnoreResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores/{itemID} [get]
func (h *Handler) GetVulnIgnore(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}

// CreateVulnIgnore Creates a vulnerability ignore
// @Summary Creates a vulnerability ignore
// @Description Creates a vulnerability ignore
// @Tags    Project env settings
// @Produce json
// @Id      createProjectEnvVulnIgnore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body vulnignoredto.CreateVulnIgnoreReq true "request data"
// @Success 201 {object} vulnignoredto.CreateVulnIgnoreResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores [post]
func (h *Handler) CreateVulnIgnore(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}

// UpdateVulnIgnore Updates vulnerability ignore
// @Summary Updates vulnerability ignore
// @Description Updates vulnerability ignore
// @Tags    Project env settings
// @Produce json
// @Id      updateProjectEnvVulnIgnore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body vulnignoredto.UpdateVulnIgnoreReq true "request data"
// @Success 200 {object} vulnignoredto.UpdateVulnIgnoreResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores/{itemID} [put]
func (h *Handler) UpdateVulnIgnore(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}

// UpdateVulnIgnoreStatus Updates vulnerability ignore status
// @Summary Updates vulnerability ignore status
// @Description Updates vulnerability ignore status
// @Tags    Project env settings
// @Produce json
// @Id      updateProjectEnvVulnIgnoreStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body vulnignoredto.UpdateVulnIgnoreStatusReq true "request data"
// @Success 200 {object} vulnignoredto.UpdateVulnIgnoreStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores/{itemID}/status [put]
func (h *Handler) UpdateVulnIgnoreStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}

// DeleteVulnIgnore Deletes vulnerability ignore
// @Summary Deletes vulnerability ignore
// @Description Deletes vulnerability ignore
// @Tags    Project env settings
// @Produce json
// @Id      deleteProjectEnvVulnIgnore
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} vulnignoredto.DeleteVulnIgnoreResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/vulnerability-ignores/{itemID} [delete]
func (h *Handler) DeleteVulnIgnore(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeVulnIgnore, base.ObjectScopeProjectEnv)
}
