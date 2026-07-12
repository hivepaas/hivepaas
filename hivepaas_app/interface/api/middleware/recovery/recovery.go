package recovery

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
)

// Recovery create a middleware for recovering from panic
func Recovery(cfg *config.Config) gin.HandlerFunc {
	// In production, use `nil` as writer to prevent Gin to log sensitive information
	// of the request to the default stderr.
	var writer io.Writer
	if cfg.IsDevEnv() {
		writer = gin.DefaultErrorWriter
	}

	return gin.CustomRecoveryWithWriter(writer, func(ctx *gin.Context, recover any) {
		err := apperrors.Wrap(apperrors.ErrInternal).
			WithMsgLog("recovered from panic: %v", recover)
		(&handler.BaseHandler{}).RenderError(ctx, err)
		ctx.Abort()
	})
}
