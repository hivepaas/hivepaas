package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// TestAFlowAgainstARealServer does, through the tools and against a running
// HivePaaS, what an assistant does for a person running an image of their own:
// it creates an app, gives it an image, deploys it and watches it start, then
// gives it a scheduled job and runs it now. It changes the server - an app and
// a job are left behind, named after the run - so it runs only when asked:
//
//	HP_TEST_MCP_FLOW   1, beside HP_TEST_MCP_URL and HP_TEST_MCP_KEY
//	HP_TEST_MCP_ENV    where to create the app, as <project>/<env>
//	HP_TEST_MCP_IMAGE  the image to run; crccheck/hello-world:latest when not set
//
// The server must allow changes, and the key must read, execute and write.
func TestAFlowAgainstARealServer(t *testing.T) {
	if os.Getenv("HP_TEST_MCP_FLOW") != "1" {
		t.Skip("set HP_TEST_MCP_FLOW=1, HP_TEST_MCP_URL, HP_TEST_MCP_KEY and HP_TEST_MCP_ENV to run it")
	}
	place := strings.Split(os.Getenv("HP_TEST_MCP_ENV"), "/")
	if !assert.Len(t, place, 2, "HP_TEST_MCP_ENV is <project>/<env>") {
		t.FailNow()
	}
	image := os.Getenv("HP_TEST_MCP_IMAGE")
	if image == "" {
		image = "crccheck/hello-world:latest"
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "hivepaas-flow-test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint:   os.Getenv("HP_TEST_MCP_URL"),
		HTTPClient: &http.Client{Transport: bearerTransport{token: os.Getenv("HP_TEST_MCP_KEY")}},
	}, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() { _ = session.Close() }()
	f := &flow{t: t, session: session}

	name := fmt.Sprintf("mcp-flow-%d", time.Now().Unix())
	inEnv := map[string]any{"project": place[0], "env": place[1]}
	inApp := merge(inEnv, map[string]any{"app": name})

	created := f.planAndApply("plan_create_app", merge(inEnv, map[string]any{"name": name, "note": "made by " +
		"TestAFlowAgainstARealServer"}))
	t.Logf("created %s: %s", name, created)

	f.planAndApply("plan_update_app_settings", merge(inApp, map[string]any{"kind": "deployment",
		"changes": map[string]any{"activeMethod": "image", "imageSource": map[string]any{"image": image}}}))

	deployed := f.planAndApply("plan_redeploy_app", inApp)
	var deploy struct {
		DeploymentID string `json:"deploymentId"`
	}
	assert.NoError(t, json.Unmarshal(deployed, &deploy))
	status := f.waitFor("get_app_deployment", merge(inApp, map[string]any{"deployment": deploy.DeploymentID}),
		"done", "failed", "canceled")
	if !assert.Equal(t, "done", status, "the deployment") {
		t.FailNow()
	}
	// A deployment is done once the swarm has the new spec; its container then
	// starts, and passes the image's health check, on its own time.
	f.waitForRunning(inApp)

	jobName := name + "-hello"
	f.planAndApply("plan_create_sched_job", merge(inApp, map[string]any{"name": jobName,
		"schedule": map[string]any{"cronExpr": "0 3 * * *"},
		"command":  map[string]any{"command": "echo hello from the MCP flow test"}}))

	ran := f.planAndApply("plan_run_sched_job", merge(inApp, map[string]any{"job": jobName}))
	var run struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	assert.NoError(t, json.Unmarshal(ran, &run))
	if assert.NotEmpty(t, run.Task.ID) {
		status = f.waitFor("get_task", merge(inApp, map[string]any{"task": run.Task.ID}), "done", "failed",
			"canceled")
		assert.Equal(t, "done", status, "the job's run")
		t.Logf("job run: %s", f.read("get_task_logs", merge(inApp, map[string]any{"task": run.Task.ID})))
	}
}

type flow struct {
	t       *testing.T
	session *mcpsdk.ClientSession
}

func (f *flow) call(tool string, args map[string]any) (string, bool) {
	f.t.Helper()
	res, err := f.session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: tool, Arguments: args})
	if !assert.NoError(f.t, err, tool) {
		f.t.FailNow()
	}
	return res.Content[0].(*mcpsdk.TextContent).Text, res.IsError
}

func (f *flow) read(tool string, args map[string]any) string {
	f.t.Helper()
	text, isErr := f.call(tool, args)
	if isErr {
		f.t.Fatalf("%s: %s", tool, text)
	}
	return text
}

// planAndApply plans, applies the plan as a person who agreed would, and
// answers what the endpoint did.
func (f *flow) planAndApply(tool string, args map[string]any) json.RawMessage {
	f.t.Helper()
	var plan planResult[json.RawMessage]
	if err := json.Unmarshal([]byte(f.read(tool, args)), &plan); err != nil || plan.PlanToken == "" {
		f.t.Fatalf("%s: no plan to apply: %v %s", tool, err, plan.Plan)
	}
	var applied applyAnswer
	raw := f.read("apply_plan", map[string]any{"planToken": plan.PlanToken})
	if err := json.Unmarshal([]byte(raw), &applied); err != nil {
		f.t.Fatalf("apply_plan of %s: %v", tool, err)
	}
	f.t.Logf("%s applied: %s", tool, applied.Summary)
	result, _ := json.Marshal(applied.Result)
	return result
}

// waitFor reads a deployment or a task until its status is one of final.
func (f *flow) waitFor(tool string, args map[string]any, final ...string) string {
	f.t.Helper()
	deadline := time.Now().Add(4 * time.Minute)
	for {
		var answer struct {
			Data struct {
				Status string `json:"status"`
			} `json:"data"`
		}
		_ = json.Unmarshal([]byte(f.read(tool, args)), &answer)
		for _, s := range final {
			if answer.Data.Status == s {
				return s
			}
		}
		if time.Now().After(deadline) {
			f.t.Fatalf("%s: still %q after 4m", tool, answer.Data.Status)
		}
		time.Sleep(3 * time.Second)
	}
}

// waitForRunning reads the app's containers until one runs.
func (f *flow) waitForRunning(inApp map[string]any) {
	f.t.Helper()
	deadline := time.Now().Add(3 * time.Minute)
	for {
		var answer struct {
			Data []struct {
				Status struct {
					State string `json:"state"`
				} `json:"status"`
			} `json:"data"`
		}
		_ = json.Unmarshal([]byte(f.read("get_app_status", inApp)), &answer)
		for _, task := range answer.Data {
			if task.Status.State == "running" {
				return
			}
		}
		if time.Now().After(deadline) {
			f.t.Fatalf("no container of the app runs after 3m")
		}
		time.Sleep(3 * time.Second)
	}
}
