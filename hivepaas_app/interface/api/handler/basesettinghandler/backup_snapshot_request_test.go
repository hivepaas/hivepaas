package basesettinghandler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
)

// A snapshot list's filters read from the query: several repositories, apps and
// tags, a date range, a search and a page; a sort asked for is dropped - newest
// first, always.
func TestBackupSnapshotListRequestParses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler.BaseHandler{}
	req := backupsnapshotdto.NewListBackupSnapshotReq()
	var parseErr error

	engine := gin.New()
	engine.GET("/", func(c *gin.Context) { parseErr = h.ParseAndValidateRequest(c, req, req.PagingReq()) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet,
		"/?repo=01JAB9XED0GTXBSQDFVYAJ8WS1&repo=01JAB9XED0GTXBSQDFVYAJ8WS2&app=01M3HT2P22YRG4PE1GGSDABGDF"+
			"&tag=env:prod&tag=db:main&fromDate=2026-09-01&toDate=2026-09-28&search=%20nightly%20"+
			"&pageOffset=20&pageLimit=10&sort=-name", nil))

	assert.NoError(t, parseErr)
	assert.Equal(t, []string{"01JAB9XED0GTXBSQDFVYAJ8WS1", "01JAB9XED0GTXBSQDFVYAJ8WS2"}, req.RepoIDs)
	assert.Equal(t, []string{"01M3HT2P22YRG4PE1GGSDABGDF"}, req.AppIDs)
	assert.Equal(t, []string{"env:prod", "db:main"}, req.Tags)
	assert.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), req.FromDate.ToTime().UTC())
	assert.Equal(t, "nightly", req.Search)
	assert.Equal(t, 20, req.Paging.Offset)
	assert.Equal(t, 10, req.Paging.Limit)
	assert.Empty(t, req.Paging.Sort)
}

// A tag without a key is refused.
func TestBackupSnapshotListRequestRefusesABadTag(t *testing.T) {
	req := backupsnapshotdto.NewListBackupSnapshotReq()
	req.Tags = []string{"prod"}
	assert.NotEmpty(t, req.Validate())
}
