package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// GetImageScanSettings Gets image scanning settings
// @Summary Gets image scanning settings
// @Description Gets image scanning's settings - the addon and its daily scan - and what runs: Trivy's server,
// @Description the runner on the nodes, and the vulnerability DB of the last scan.
// @Tags    System settings
// @Produce json
// @Id      getSystemImageScanSettings
// @Success 200 {object} imagescandto.GetImageScanSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan [get]
// openapi:ignore-param revealSecrets - the settings hold no secret
func (h *Handler) GetImageScanSettings(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewGetImageScanSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.GetImageScanSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateImageScanSettings Updates image scanning settings
// @Summary Updates image scanning settings
// @Description Switches image scanning on - Trivy's server and a runner on every node are made - or off -
// @Description both are removed, the results kept - and sets its daily scan.
// @Tags    System settings
// @Produce json
// @Id      updateSystemImageScanSettings
// @Param   body body imagescandto.UpdateImageScanSettingsReq true "request data"
// @Success 200 {object} imagescandto.UpdateImageScanSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan [put]
func (h *Handler) UpdateImageScanSettings(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewUpdateImageScanSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.UpdateImageScanSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteImageScan Runs the daily image scan now
// @Summary Runs the daily image scan now
// @Description Scans again what every app runs, as the daily scan does.
// @Tags    System settings
// @Produce json
// @Id      executeSystemImageScan
// @Success 200 {object} imagescandto.ExecuteImageScanResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan/exec [post]
// openapi:ignore-param revealSecrets - a run reveals nothing
func (h *Handler) ExecuteImageScan(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeExecute)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewExecuteImageScanReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.ExecuteImageScan(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ListImageScanApps Lists every app's images and what they carry
// @Summary Lists every app's images and what they carry
// @Description Every scanned app with its images and their findings counted by severity, after ignores; the
// @Description worst first, by fixable CRITICAL then fixable HIGH.
// @Tags    System settings
// @Produce json
// @Id      listSystemImageScanApps
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Success 200 {object} imagescandto.ListImageScanAppsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan/apps [get]
func (h *Handler) ListImageScanApps(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewListImageScanAppsReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.ListImageScanApps(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) imageScanAuth(ctx *gin.Context, action base.ActionType) (*basedto.Auth, error) {
	return h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{ //nolint:wrapcheck
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeImageScan,
	})
}

// ListImageScanComponents Lists what HivePaaS's own parts run and carry
// @Summary Lists what HivePaaS's own parts run and carry
// @Description HivePaaS's parts the daily scan found running - the app, the agent, the database, Redis,
// @Description Traefik, the registry, the logging backend and collector, OBI - each with its images and their
// @Description findings counted by severity, after the global ignores, in HivePaaS's order.
// @Tags    System settings
// @Produce json
// @Id      listSystemImageScanComponents
// @Success 200 {object} imagescandto.ListImageScanComponentsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan/hivepaas [get]
func (h *Handler) ListImageScanComponents(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewListImageScanComponentsReq()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.ListImageScanComponents(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetImageScanComponent Gets what a part of HivePaaS runs and carries
// @Summary Gets what a part of HivePaaS runs and carries
// @Description The images the part ran when last scanned, its findings no global ignore sets aside, the worst
// @Description first, and those one does, each with its ignore.
// @Tags    System settings
// @Produce json
// @Id      getSystemImageScanComponent
// @Param   component path string true "the part's key: app, agent, db, redis, traefik, registry, vlagent…"
// @Success 200 {object} imagescandto.GetImageScanComponentResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/image-scan/hivepaas/{component} [get]
func (h *Handler) GetImageScanComponent(ctx *gin.Context) {
	auth, err := h.imageScanAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagescandto.NewGetImageScanComponentReq()
	req.Component = ctx.Param("component")
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageScanUC.GetImageScanComponent(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
