package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerSessionRoutes(apiGroup *gin.RouterGroup) {
	sessionGroup := apiGroup.Group("/sessions")
	sessionHandler := s.handlerRegistry.sessionHandler

	// User info
	sessionGroup.GET("/me", sessionHandler.GetMe)
	// Session handling
	sessionGroup.POST("/refresh", sessionHandler.RefreshSession)
	sessionGroup.DELETE("", sessionHandler.DeleteSession)
	sessionGroup.POST("/delete-all", sessionHandler.DeleteAllSessions)

	authGroup := apiGroup.Group("/auth")
	{
		// Login options
		authGroup.GET("/login-options", sessionHandler.LoginGetOptions)
		// Login with password
		authGroup.POST("/login-with-password", sessionHandler.LoginWithPassword)
		authGroup.POST("/login-with-passcode", sessionHandler.LoginWithPasscode)
		// Login with API key
		authGroup.POST("/login-with-api-key", sessionHandler.LoginWithAPIKey)
		// Login via SSO
		authGroup.GET("/sso/:provider", sessionHandler.SSOOAuthBegin)
		authGroup.GET("/sso/callback/:provider", sessionHandler.SSOOAuthCallback)
		authGroup.POST("/sso/callback/:provider", sessionHandler.SSOOAuthCallback)
		// Password forgot
		authGroup.POST("/login-password-forgot", sessionHandler.LoginPasswordForgot)
	}
}
