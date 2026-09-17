package server

import (
	"github.com/gin-gonic/gin"
)

// registerAppTemplateRoutes registers the public icon route. The catalog itself
// is read under a project - see registerProjectRoutes - and creating an app from
// a template under its env - see registerAppRoutes.
func (s *HTTPServer) registerAppTemplateRoutes(apiGroup *gin.RouterGroup) {
	apiGroup.GET("/app-templates/icons/:sha256", s.handlerRegistry.appTemplateHandler.GetAppTemplateIcon)
}
