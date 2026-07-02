package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerImageRoutes(apiGroup *gin.RouterGroup) {
	imageGroup := apiGroup.Group("/images")
	imageHandler := s.handlerRegistry.imageHandler

	imageGroup.GET("/:imageID", imageHandler.GetPublicImage)
}
