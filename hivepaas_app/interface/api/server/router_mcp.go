package server

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/mcp"
)

// registerMCPRoutes serves the Model Context Protocol at <base>/mcp. The
// endpoint's tools dispatch to this same engine, which is why it is built here,
// where the engine is.
func (s *HTTPServer) registerMCPRoutes(apiGroup *gin.RouterGroup) {
	dispatcher := mcp.NewDispatcher(s.engine, s.config.HTTPServer.BasePath)
	endpoint := mcp.NewEndpoint(s.mcpServices, dispatcher, mcp.Tools())
	apiGroup.Any("/mcp", endpoint.Serve)
}
