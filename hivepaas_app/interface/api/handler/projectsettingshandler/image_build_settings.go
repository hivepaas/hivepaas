package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

// GetImageBuildSettings Gets image build setting details
// @Summary Gets image build setting details
// @Description Gets image build setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} imagebuildsettingsdto.GetImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [get]
func (h *Handler) GetImageBuildSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeProject)
}

// UpdateImageBuildSettings Updates image build settings
// @Summary Updates image build settings
// @Description Updates image build settings
// @Tags    project_settings
// @Produce json
// @Id      updateProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [put]
func (h *Handler) UpdateImageBuildSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeProject)
}

// UpdateImageBuildSettingsStatus Updates image build status
// @Summary Updates image build status
// @Description Updates image build status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectImageBuildSettingsStatus
// @Param   projectID path string true "project ID"
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsStatusReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings/status [put]
func (h *Handler) UpdateImageBuildSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeProject)
}

// DeleteImageBuildSettings Deletes image build settings
// @Summary Deletes image build settings
// @Description Deletes image build settings
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} imagebuildsettingsdto.DeleteImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [delete]
func (h *Handler) DeleteImageBuildSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeProject)
}

// GetRepoCacheInfo Gets repo cache info
// @Summary Gets repo cache info
// @Description Gets repo cache info
// @Tags    project_settings
// @Produce json
// @Id      getProjectRepoCacheInfo
// @Param   projectID path string true "project ID"
// @Success 200 {object} imagebuildsettingsdto.GetRepoCacheInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings/repo-cache [get]
func (h *Handler) GetRepoCacheInfo(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagebuildsettingsdto.NewGetRepoCacheInfoReq()
	req.Scope = base.NewObjectScopeProject(projectID)
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageBuildUC.GetRepoCacheInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ClearRepoCache Clears repo cache
// @Summary Clears repo cache
// @Description Clears repo cache
// @Tags    project_settings
// @Produce json
// @Id      clearProjectRepoCache
// @Param   projectID path string true "project ID"
// @Param   body body imagebuildsettingsdto.ClearRepoCacheReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.ClearRepoCacheResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings/repo-cache/clear [post]
func (h *Handler) ClearRepoCache(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeExecute, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagebuildsettingsdto.NewClearRepoCacheReq()
	req.Scope = base.NewObjectScopeProject(projectID)
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageBuildUC.ClearRepoCache(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
