// Package appimagescanhandler serves an app's image scan: what its images
// carry, and a scan asked for.
package appimagescanhandler

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/appbasehandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc"
)

type Handler struct {
	*appbasehandler.Handler
	appImageScanUC *appimagescanuc.UC
}

func New(
	baseHandler *appbasehandler.Handler,
	appImageScanUC *appimagescanuc.UC,
) *Handler {
	return &Handler{
		Handler:        baseHandler,
		appImageScanUC: appImageScanUC,
	}
}
