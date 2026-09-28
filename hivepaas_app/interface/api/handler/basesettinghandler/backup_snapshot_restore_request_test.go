package basesettinghandler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// A download names its file in the query.
func TestBackupSnapshotDownloadRequestParses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler.BaseHandler{}
	req := backupsnapshotdto.NewDownloadBackupSnapshotFileReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WS9"
	var parseErr error

	engine := gin.New()
	engine.GET("/", func(c *gin.Context) { parseErr = h.ParseAndValidateRequest(c, req, nil) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?path=spec.tar.gz.age", nil))

	assert.NoError(t, parseErr)
	assert.Equal(t, "spec.tar.gz.age", req.Path)
}

var _ = (*Handler).DownloadBackupSnapshotFile

// A download is an attachment of its own name and size.
func TestDownloadHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	writeDownloadHeaders(ctx, &backupsnapshotdto.DownloadBackupSnapshotFileResp{
		FileName: "spec.tar.gz.age", SizeBytes: 42,
	})

	assert.Equal(t, `attachment; filename=spec.tar.gz.age`, rec.Header().Get("Content-Disposition"))
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "42", rec.Header().Get("Content-Length"))
}

// A download outlasts the server's write timeout: a large file on a slow link
// takes minutes, and the timeout is meant for ordinary answers.
func TestADownloadOutlastsTheWriteTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chunk := strings.Repeat("x", 1024)
	const chunks = 4
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		writeDownload(c, c.Request.Context(), &backupsnapshotdto.DownloadBackupSnapshotFileResp{
			FileName: "db.pg_dump", SizeBytes: int64(len(chunk) * chunks),
			Write: func(_ context.Context, w io.Writer) error {
				for range chunks {
					time.Sleep(100 * time.Millisecond)
					if _, err := io.WriteString(w, chunk); err != nil {
						return err
					}
					w.(http.Flusher).Flush()
				}
				return nil
			},
		})
	})
	server := httptest.NewUnstartedServer(engine)
	server.Config.WriteTimeout = 150 * time.Millisecond
	server.Start()
	defer server.Close()

	resp, err := http.Get(server.URL) //nolint:noctx
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Len(t, body, len(chunk)*chunks)
}
