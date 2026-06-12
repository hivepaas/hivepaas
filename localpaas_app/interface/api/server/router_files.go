package server

import (
	"github.com/gin-gonic/gin"
)

func (s *HTTPServer) registerFileRoutes(apiGroup *gin.RouterGroup) {
	fileGroup := apiGroup.Group("/files")
	fileHandler := s.handlerRegistry.fileHandler

	fileGroup.GET("/:fileID/download", fileHandler.DownloadFile)
	fileGroup.POST("/upload", fileHandler.UploadFiles)
}
