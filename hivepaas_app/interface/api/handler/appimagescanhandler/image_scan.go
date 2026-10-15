package appimagescanhandler

import (
	"context"
	"mime"
	"net/http"
	"time"

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

// sbomWriteTimeout lifts the server's write timeout for an SBOM: Trivy reads
// the image, which a cold image or DB takes minutes to; the wait for the
// volume is half a minute.
const sbomWriteTimeout = 15 * time.Minute

// GetAppImageSBOM Downloads the SBOM of an image an app runs
// @Summary Downloads the SBOM of an image an app runs
// @Description The image's packages as CycloneDX JSON, read as a scan reads the image - on the same node,
// @Description with the same volume - and not kept: a few seconds once the image was scanned. Only an image
// @Description the app's tasks run now.
// @Tags    Apps
// @Produce application/vnd.cyclonedx+json
// @Id      getAppImageSBOM
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   imageId query string true "the image's ID, as the app's Security tab lists it"
// @Success 200 {file} binary
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/image-scan/sbom [get]
func (h *Handler) GetAppImageSBOM(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appimagescandto.NewGetAppImageSBOMReq()
	req.ProjectID, req.ProjectEnvID, req.AppID = projectID, projectEnvID, appID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = http.NewResponseController(ctx.Writer).SetWriteDeadline(time.Now().Add(sbomWriteTimeout)); err != nil {
		_ = ctx.Error(err)
	}
	// The SBOM's scan ends with its request: one whose person went away stops,
	// and its container is removed, rather than holding the volume.
	reqCtx, cancel := context.WithCancel(h.RequestCtx(ctx))
	defer cancel()
	defer context.AfterFunc(ctx.Request.Context(), cancel)()

	resp, err := h.appImageScanUC.GetAppImageSBOM(reqCtx, auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": resp.FileName}))
	ctx.Data(http.StatusOK, "application/vnd.cyclonedx+json", resp.Content)
}
