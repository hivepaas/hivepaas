package obiagentuc

import (
	"context"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/services/docker"
)

// Run against the Docker this machine runs, with HP_TEST_OBI_DOCKER=1: a node
// whose kernel OBI can probe (Docker Desktop's can). It starts OBI as the agent
// does - its options, its configuration copied in, in a stand-in agent's
// network namespace - and checks that it counts the requests to an app whose
// containers are named as swarm names a task's.
func TestLiveEnsureRunsOBIThatCountsTheApp(t *testing.T) {
	if os.Getenv("HP_TEST_OBI_DOCKER") == "" {
		t.Skip("HP_TEST_OBI_DOCKER not set")
	}
	ctx := context.Background()
	dm, err := docker.New()
	if err != nil {
		t.Fatal(err)
	}
	const network, agent, app = "hp-obi-live", "hp-obi-live-agent", "hp_obi_live.1.abc123"
	cleanup := func() {
		for _, name := range []string{obi.ContainerName, agent, app} {
			_, _ = dm.ContainerRemove(ctx, name, func(o *client.ContainerRemoveOptions) { o.Force = true })
		}
		_, _ = dm.NetworkRemove(ctx, network)
	}
	cleanup()
	t.Cleanup(cleanup)

	if _, err = dm.NetworkCreate(ctx, network); err != nil {
		t.Fatal(err)
	}
	run := func(name, image string, cmd []string) string {
		created, err := dm.ContainerCreate(ctx, func(o *client.ContainerCreateOptions) {
			o.Name = name
			o.Config = &container.Config{Image: image, Cmd: cmd}
			o.HostConfig = &container.HostConfig{NetworkMode: container.NetworkMode(network)}
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = dm.ContainerStart(ctx, created.ID); err != nil {
			t.Fatal(err)
		}
		return created.ID
	}
	agentID := run(agent, "busybox:latest", []string{"sleep", "600"})
	run(app, liveAppImage, []string{"node", "-e",
		"require('http').createServer((q,s)=>{s.end('ok')}).listen(8080)"})

	uc := New(quiet{}, nil, nil, nil, dm, "")
	uc.agentID = agentID
	assert.NoError(t, uc.ensure(ctx, obi.Config(obi.Patterns([]string{"hp_obi_live"}))))
	assert.True(t, uc.running)

	exec := func(cmd string) string {
		_, frames, err := dm.ContainerExecWait(ctx, agentID, func(o *client.ExecCreateOptions) {
			o.Cmd, o.AttachStdout, o.AttachStderr = []string{"sh", "-c", cmd}, true, true
		})
		if err != nil {
			t.Fatal(err)
		}
		var out strings.Builder
		for _, f := range frames {
			out.WriteString(f.Data + "\n")
		}
		return out.String()
	}
	// OBI serves its metrics in the agent's network namespace: on localhost.
	var metrics string
	for range 30 {
		time.Sleep(2 * time.Second)
		if metrics = exec("wget -qO- http://127.0.0.1:" + strconv.Itoa(obi.MetricsPort) + "/metrics"); metrics != "" {
			break
		}
	}
	if !assert.NotEmpty(t, metrics, "OBI serves its metrics on the agent's localhost") {
		return
	}
	time.Sleep(10 * time.Second) // OBI attaches to the app it found
	exec("for i in $(seq 1 50); do wget -qO- http://" + app + ":8080/x/ >/dev/null; done")
	count := regexp.MustCompile(`http_server_request_duration_seconds_count\{[^}]*container_name="` +
		regexp.QuoteMeta(app) + `"[^}]*\} ([0-9]+)`)
	var got int
	for range 15 {
		time.Sleep(2 * time.Second)
		if m := count.FindStringSubmatch(exec("wget -qO- http://127.0.0.1:" + strconv.Itoa(obi.MetricsPort) +
			"/metrics")); m != nil {
			if got, _ = strconv.Atoi(m[1]); got >= 50 {
				break
			}
		}
	}
	if !assert.GreaterOrEqual(t, got, 50, "the app's requests, counted by its container's name") {
		logs, _ := dm.ContainerLogs(ctx, obi.ContainerName, func(o *client.ContainerLogsOptions) {
			o.ShowStdout, o.ShowStderr, o.Tail = true, true, "30"
		})
		if logs != nil {
			b, _ := io.ReadAll(logs)
			t.Logf("OBI's logs:\n%s", b)
		}
	}
}

// liveAppImage is the app the live test loads: a Node.js server, as step 0's.
const liveAppImage = "node:24-trixie-slim"
