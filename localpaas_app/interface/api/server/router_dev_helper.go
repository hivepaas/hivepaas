package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerDevRoutes(apiGroup *gin.RouterGroup) *gin.RouterGroup {
	if !s.config.IsDevEnv() {
		return nil
	}
	devHelperGroup := apiGroup.Group("/dev-helper")
	devHelperHandler := s.handlerRegistry.devHelperHandler

	devHelperGroup.POST("/lock-task", devHelperHandler.LockTask)
	devHelperGroup.POST("/long-req", devHelperHandler.SimulateLongRequest)
	devHelperGroup.POST("/exec-cmd", devHelperHandler.ExecuteCmd)

	return devHelperGroup
}
