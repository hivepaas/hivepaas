package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerUserRoutes(apiGroup *gin.RouterGroup) (*gin.RouterGroup, *gin.RouterGroup) {
	userGroup := apiGroup.Group("/users")
	userHandler := s.handlerRegistry.userHandler

	{ // user group
		// User info
		userGroup.GET("/base", userHandler.ListUserBase)
		userGroup.GET("/:userID", userHandler.GetUser)
		userGroup.GET("", userHandler.ListUser)
		// Password
		userGroup.PUT("/current/password", userHandler.UpdateUserPassword)
		userGroup.POST("/:userID/password/request-reset", userHandler.RequestResetPassword)
		userGroup.POST("/:userID/password/reset", userHandler.ResetPassword)
		// Profile
		userGroup.PUT("/current/profile", userHandler.UpdateUserProfile)
		// Update (admin API)
		userGroup.PUT("/:userID", userHandler.UpdateUser)
		userGroup.DELETE("/:userID", userHandler.DeleteUser)
		// MFA TOTP setup
		userGroup.POST("/current/mfa/totp-begin-setup", userHandler.BeginMFATotpSetup)
		userGroup.POST("/current/mfa/totp-complete-setup", userHandler.CompleteMFATotpSetup)
		userGroup.POST("/current/mfa/totp-remove", userHandler.RemoveMFATotp)
		// Invite & SignUp
		userGroup.GET("/invite/info", userHandler.GetUserInviteInfo)
		userGroup.POST("/invite", userHandler.InviteUser)
		userGroup.POST("/signup-begin", userHandler.BeginUserSignup)
		userGroup.POST("/signup-complete", userHandler.CompleteUserSignup)
	}

	// User settings group
	userSettingGroup := userGroup.Group("/current/settings")
	userSettingsHandler := s.handlerRegistry.userSettingsHandler

	{ // API key group
		apiKeyGroup := userSettingGroup.Group("/api-keys")
		apiKeyGroup.GET("/:itemID", userSettingsHandler.GetAPIKey)
		apiKeyGroup.GET("", userSettingsHandler.ListAPIKey)
		apiKeyGroup.POST("", userSettingsHandler.CreateAPIKey)
		apiKeyGroup.PUT("/:itemID/meta", userSettingsHandler.UpdateAPIKeyMeta)
		apiKeyGroup.DELETE("/:itemID", userSettingsHandler.DeleteAPIKey)
	}

	return userGroup, userSettingGroup
}
