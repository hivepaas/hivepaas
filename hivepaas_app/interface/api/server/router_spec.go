package server

import (
	"github.com/gin-gonic/gin"
)

// registerSpecRoutes registers the global-scope export. The narrower scopes are
// registered beside the routes they belong to - see router_projects.go,
// router_project_env.go and router_apps.go - so that a reader of those files
// sees everything reachable under that prefix.
func (s *HTTPServer) registerSpecRoutes(apiGroup *gin.RouterGroup) {
	apiGroup.GET("/spec/export", s.handlerRegistry.specHandler.ExportGlobalSpec)
}
