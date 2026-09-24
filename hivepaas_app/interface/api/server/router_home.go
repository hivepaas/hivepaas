package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerHomeRoutes(apiGroup *gin.RouterGroup) {
	homeGroup := apiGroup.Group("/home")
	homeHandler := s.handlerRegistry.homeHandler

	homeGroup.GET("/attention", homeHandler.GetHomeAttention)
}
