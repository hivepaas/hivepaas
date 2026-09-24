package dockerproxy

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTheAppsChildrenAreReachable(t *testing.T) {
	w := newWorld(t, testPolicy())
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodGet, "/v1.51/containers/child1/json"},
		{http.MethodGet, "/v1.51/containers/child1/logs?stdout=1"},
		{http.MethodPost, "/v1.51/containers/child1/start"},
		{http.MethodPost, "/v1.51/containers/child1/wait"},
		{http.MethodDelete, "/v1.51/containers/child1?force=1"},
	} {
		status, raw := w.do(t, endpoint.method, endpoint.path, nil)
		assert.Equal(t, http.StatusOK, status, "%s %s", endpoint.path, raw)
	}
}

func TestOtherContainersAreNot(t *testing.T) {
	w := newWorld(t, testPolicy())
	tests := map[string]string{
		// Another app's child, and the app's own task: the app may run containers,
		// not reach into the ones HivePaaS runs.
		"/v1.51/containers/other1/json":  "container other1 is not one this app started",
		"/v1.51/containers/task1/json":   "container task1 is not one this app started",
		"/v1.51/containers/missing/json": "container missing does not exist",
	}
	for path, message := range tests {
		status, raw := w.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusForbidden, status, path)
		assert.Equal(t, "hivepaas: "+message, refusalMessage(t, raw), path)
	}
	status, _ := w.do(t, http.MethodDelete, "/v1.51/containers/other1", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.False(t, w.reached(http.MethodDelete, "/containers/other1"))
}

func TestTheContainerListShowsOnlyTheAppsChildren(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodGet, "/v1.51/containers/json?all=1", nil)
	stop(t, assert.Equal(t, http.StatusOK, status))
	var containers []struct {
		ID string `json:"Id"`
	}
	stop(t, assert.NoError(t, json.Unmarshal(raw, &containers)))
	stop(t, assert.Len(t, containers, 1))
	assert.Equal(t, "child1", containers[0].ID)
}

func TestAttachStreamsThroughTheProxy(t *testing.T) {
	w := newWorld(t, testPolicy())
	conn, err := net.Dial("tcp", strings.TrimPrefix(w.url, "http://"))
	stop(t, assert.NoError(t, err))
	defer conn.Close()
	_, err = conn.Write([]byte("POST /v1.51/containers/child1/attach?stream=1&stdout=1 HTTP/1.1\r\n" +
		"Host: docker\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n"))
	stop(t, assert.NoError(t, err))
	stop(t, assert.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second))))
	got, _ := io.ReadAll(conn)
	assert.Contains(t, string(got), "101")
	assert.Contains(t, string(got), "stream-ok")
}

func TestGroupsThePolicyDoesNotAllowAreRefused(t *testing.T) {
	w := newWorld(t, testPolicy())
	status, raw := w.do(t, http.MethodPost, "/v1.51/containers/child1/exec", map[string]any{"Cmd": []string{"sh"}})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec is not allowed for this app", refusalMessage(t, raw))

	status, raw = w.do(t, http.MethodGet, "/v1.51/containers/child1/archive?path=/", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: files is not allowed for this app", refusalMessage(t, raw))
}

func TestExecRunsOnlyInTheAppsChildren(t *testing.T) {
	policy := testPolicy()
	policy.Allow = []Group{GroupExec, GroupFiles}
	w := newWorld(t, policy)

	status, raw := w.do(t, http.MethodPost, "/v1.51/containers/child1/exec",
		map[string]any{"Cmd": []string{"sh", "-c", "id"}, "AttachStdout": true, "Privileged": false})
	assert.Equal(t, http.StatusOK, status, string(raw))

	status, raw = w.do(t, http.MethodPost, "/v1.51/containers/child1/exec",
		map[string]any{"Cmd": []string{"sh"}, "Privileged": true})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec.Privileged is not allowed", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodPost, "/v1.51/containers/other1/exec", map[string]any{"Cmd": []string{"sh"}})
	assert.Equal(t, http.StatusForbidden, status)

	status, _ = w.do(t, http.MethodPost, "/v1.51/exec/exec-child/start", map[string]any{"Detach": false})
	assert.Equal(t, http.StatusOK, status)
	status, raw = w.do(t, http.MethodPost, "/v1.51/exec/exec-other/start", map[string]any{"Detach": false})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: container other1 is not one this app started", refusalMessage(t, raw))
	status, raw = w.do(t, http.MethodGet, "/v1.51/exec/missing/json", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "hivepaas: exec missing does not exist", refusalMessage(t, raw))

	status, _ = w.do(t, http.MethodPut, "/v1.51/containers/child1/archive?path=/tmp", []byte("tar"))
	assert.Equal(t, http.StatusOK, status)
}
