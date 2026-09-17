package server

import (
	"github.com/gin-gonic/gin"
)

// registerAppTemplateRoutes registers the public icon route. The catalog itself
// is read under a project - see registerProjectRoutes - and creating an app from
// a template under its env - see registerAppRoutes.
func (s *HTTPServer) registerAppTemplateRoutes(apiGroup *gin.RouterGroup) {
	appTemplateGroup := apiGroup.Group("/app-templates")
	appTemplateHandler := s.handlerRegistry.appTemplateHandler

	appTemplateGroup.GET("/catalog", appTemplateHandler.GetAppTemplateCatalog)
	appTemplateGroup.GET("", appTemplateHandler.ListAppTemplates)
	appTemplateGroup.GET("/:templateName", appTemplateHandler.GetAppTemplate)
	appTemplateGroup.GET("/:templateName/image-tags", appTemplateHandler.GetAppTemplateImageTags)
	appTemplateGroup.GET("/icons/:file", appTemplateHandler.GetAppTemplateIcon)
}
