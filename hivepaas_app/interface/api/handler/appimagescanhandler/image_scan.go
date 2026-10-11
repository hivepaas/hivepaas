package appimagescanhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// GetAppImageScan Gets what an app's images carry
// @Summary Gets what an app's images carry
// @Description The images the app's tasks ran when last scanned, each with its last scan; the findings no
// @Description ignore of the app's scopes sets aside, the worst first; those one does, each with its ignore;
// @Description and the app's scan queued or running, if any.
// @Tags    Apps
// @Produce json
// @Id      getAppImageScan
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appimagescandto.GetAppImageScanResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/image-scan [get]
func (h *Handler) GetAppImageScan(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appimagescandto.NewGetAppImageScanReq()
	req.ProjectID, req.ProjectEnvID, req.AppID = projectID, projectEnvID, appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appImageScanUC.GetAppImageScan(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ScanAppImages Scans an app's images now
// @Summary Scans an app's images now
// @Description Queues a scan of the images the app's tasks run, scanning again those scanned within the hour.
// @Description A scan of the app queued before and not started yet is the one answered.
// @Tags    Apps
// @Produce json
// @Id      scanAppImages
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} appimagescandto.ScanAppImagesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/image-scan [post]
func (h *Handler) ScanAppImages(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appimagescandto.NewScanAppImagesReq()
	req.ProjectID, req.ProjectEnvID, req.AppID = projectID, projectEnvID, appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appImageScanUC.ScanAppImages(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
