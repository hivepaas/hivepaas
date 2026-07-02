package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSupportRoutes(apiGroup *gin.RouterGroup) {
	supportGroup := apiGroup.Group("/support")
	supportHandler := s.handlerRegistry.supportHandler

	{ // Feedback group
		feedbackGroup := supportGroup.Group("/feedbacks")
		feedbackGroup.POST("", supportHandler.CreateFeedback)
	}
}
