package basesettinghandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// A directory of a snapshot is asked for by its path, in the query.
func TestBackupSnapshotEntriesRequestParses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler.BaseHandler{}
	req := backupsnapshotdto.NewListBackupSnapshotEntriesReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WS9"
	var parseErr error

	engine := gin.New()
	engine.GET("/", func(c *gin.Context) { parseErr = h.ParseAndValidateRequest(c, req, nil) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?path=uploads/2026/", nil))

	assert.NoError(t, parseErr)
	assert.Equal(t, "uploads/2026", req.Path)
}

// A restore is asked for in the body.
func TestBackupSnapshotRestoreRequestParses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler.BaseHandler{}
	req := backupsnapshotdto.NewRestoreBackupSnapshotReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WS9"
	var parseErr error

	engine := gin.New()
	engine.POST("/", func(c *gin.Context) { parseErr = h.ParseAndValidateJSONBody(c, req) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"targetApp":{"id":"01M3HT2P22YRG4PE1GGSDABGDF"},"volume":{"id":"01M3HT2P22YRG4PE1GGSDABGD0"},`+
			`"subpath":"data","snapshotPath":"uploads","stopApp":true,"mode":"replace"}`)))

	assert.NoError(t, parseErr)
	assert.Equal(t, "01M3HT2P22YRG4PE1GGSDABGDF", req.TargetApp.ID)
	assert.Equal(t, "uploads", req.SnapshotPath)
	assert.Equal(t, base.BackupRestoreModeReplace, req.Mode)
	assert.True(t, req.StopApp)
}

var _ = (*Handler).RestoreBackupSnapshot
var _ = (*Handler).ListBackupSnapshotEntries
