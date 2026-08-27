package recovery

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
)

// Recovery create a middleware for recovering from panic
func Recovery(cfg *config.Config, baseHandler *handler.BaseHandler) gin.HandlerFunc {
	// In production, use `nil` as writer to prevent Gin to log sensitive information
	// of the request to the default stderr.
	var writer io.Writer
	if cfg.IsDevEnv() {
		writer = gin.DefaultErrorWriter
	}

	return gin.CustomRecoveryWithWriter(writer, func(ctx *gin.Context, recover any) {
		err := hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("recovered from panic: %v", recover)
		if baseHandler != nil {
			baseHandler.RenderError(ctx, err)
		} else {
			(&handler.BaseHandler{}).RenderError(ctx, err)
		}
		ctx.Abort()
	})
}
