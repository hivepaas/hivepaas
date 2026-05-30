package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerWebhookRoutes(apiGroup *gin.RouterGroup) {
	webhookGroup := apiGroup.Group("/webhooks")
	webhookHandler := s.handlerRegistry.webhookHandler

	// Repo webhook
	webhookGroup.POST("/:webhookID", webhookHandler.HandleRepoWebhook)
}
