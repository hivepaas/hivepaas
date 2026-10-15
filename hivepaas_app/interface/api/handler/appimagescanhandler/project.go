package appimagescanhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// GetProjectImageScan Gets what the images of a project's apps carry
// @Summary Gets what the images of a project's apps carry
// @Description The project's scanned apps, each with its images and their counts after the ignores of its
// @Description scopes, the worst first: by fixable CRITICAL, then fixable HIGH. Only the envs and the apps the
// @Description caller may read are listed. enabled says whether image scanning is on.
// @Tags    Projects
// @Produce json
// @Id      getProjectImageScan
// @Param   projectID path string true "project ID"
// @Success 200 {object} appimagescandto.GetProjectImageScanResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/image-scan [get]
func (h *Handler) GetProjectImageScan(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appimagescandto.NewGetProjectImageScanReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appImageScanUC.GetProjectImageScan(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
