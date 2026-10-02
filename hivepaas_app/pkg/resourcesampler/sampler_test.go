package resourcesampler

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func writeCgroup(t *testing.T, root, id string, usage string) {
	t.Helper()
	dir := filepath.Join(root, "docker", id)
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	for name, content := range map[string]string{
		"cpu.stat": "usage_usec " + usage + "\n", "cpu.max": "max 100000\n",
		"memory.current": "4096\n", "memory.max": "8192\n",
	} {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
}

// A container gets a row from its second sample on; one that went is
// forgotten; one with no cgroup is passed over.
func TestTick(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	at := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	s := &Sampler{CgroupRoot: root, ProcRoot: t.TempDir(), Out: &out, Now: func() time.Time { return at }}
	web := Container{ID: "aaaaaaaaaaaaaaaaaaaa", AppID: "01K6A"}
	gone := Container{ID: "bbbb", AppID: "01K6B"}
	ghost := Container{ID: "cccc", AppID: "01K6C"}

	writeCgroup(t, root, web.ID, "1000000")
	writeCgroup(t, root, gone.ID, "1000000")
	assert.Equal(t, 0, s.Tick(context.Background(), []Container{web, gone, ghost}), "a first sample has no row")

	at = at.Add(15 * time.Second)
	writeCgroup(t, root, web.ID, "4000000")
	assert.Equal(t, 1, s.Tick(context.Background(), []Container{web, ghost}))
	assert.NotContains(t, s.prev, gone.ID, "forgotten")

	var row Row
	assert.NoError(t, json.Unmarshal(bytes.TrimSpace(out.Bytes()), &row))
	assert.Equal(t, Row{HP: "resources", App: "01K6A", Container: "aaaaaaaaaaaa", CPU: 0.2, Memory: 4096,
		MemLimit: 8192}, row)
	assert.Equal(t, 1, strings.Count(out.String(), "\n"), "one line")

	s.Reset()
	at = at.Add(15 * time.Second)
	assert.Equal(t, 0, s.Tick(context.Background(), []Container{web}), "after a pause, a first sample again")
}
