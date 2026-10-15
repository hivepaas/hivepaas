package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestImageScanAgainstARealServer does, through the tools and against a
// running HivePaaS whose image scanning is on, what an assistant does for a
// person asking what an app's image carries: it runs alpine:3.10 - past its
// support, its apk-tools with CVE-2021-36159, CRITICAL, fixed in 2.10.7-r0 -
// scans it now and reads the finding. Like TestAFlowAgainstARealServer it
// leaves its app behind, named after the run, and runs only when asked:
//
//	HP_TEST_MCP_SCAN  1, beside HP_TEST_MCP_URL, HP_TEST_MCP_KEY and HP_TEST_MCP_ENV
//
// The server must allow changes and scan images, and the key must read,
// execute and write.
func TestImageScanAgainstARealServer(t *testing.T) {
	if os.Getenv("HP_TEST_MCP_SCAN") != "1" {
		t.Skip("set HP_TEST_MCP_SCAN=1, HP_TEST_MCP_URL, HP_TEST_MCP_KEY and HP_TEST_MCP_ENV to run it")
	}
	place := strings.Split(os.Getenv("HP_TEST_MCP_ENV"), "/")
	if !assert.Len(t, place, 2, "HP_TEST_MCP_ENV is <project>/<env>") {
		t.FailNow()
	}
	f := newFlow(t)

	name := fmt.Sprintf("mcp-scan-%d", time.Now().Unix())
	inEnv := map[string]any{"project": place[0], "env": place[1]}
	inApp := merge(inEnv, map[string]any{"app": name})

	f.planAndApply("plan_create_app", merge(inEnv, map[string]any{"name": name, "note": "made by " +
		"TestImageScanAgainstARealServer"}))
	// A task that runs is what a scan reads: alpine's own command ends at once.
	f.planAndApply("plan_update_app_settings", merge(inApp, map[string]any{"kind": "deployment",
		"changes": map[string]any{"activeMethod": "image", "imageSource": map[string]any{"image": "alpine:3.10"},
			"command": "sleep 3600"}}))
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
	f.waitForRunning(inApp)

	var plan planResult[scanAppPlan]
	if err := json.Unmarshal([]byte(f.read("plan_scan_app_images", inApp)), &plan); err != nil {
		t.Fatalf("plan_scan_app_images: %v", err)
	}
	if plan.PlanToken == "" {
		t.Fatalf("no scan to apply - the server must scan images: %v", plan.Plan.Issues)
	}
	assert.NotEmpty(t, plan.Plan.Effect)
	var scan struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	assert.NoError(t, json.Unmarshal(f.apply("plan_scan_app_images", plan.PlanToken), &scan))
	if !assert.NotEmpty(t, scan.Task.ID) {
		t.FailNow()
	}
	// The first scan on a volume fetches Trivy's databases; the deployment's
	// own scan may be ahead of this one on the volume's lock.
	status = f.waitForWithin(10*time.Minute, "get_task", merge(inApp, map[string]any{"task": scan.Task.ID}),
		"done", "failed", "canceled")
	if !assert.Equal(t, "done", status, "the scan") {
		t.FailNow()
	}

	var vulns vulnsAnswer
	assert.NoError(t, json.Unmarshal([]byte(f.read("get_app_vulnerabilities", inApp)), &vulns))
	assert.True(t, vulns.Enabled)
	assert.NotNil(t, vulns.ScannedAt)
	if assert.Len(t, vulns.Images, 1) {
		assert.Equal(t, "alpine:3.10", vulns.Images[0].ImageRef)
		assert.True(t, vulns.Images[0].OSEOSL, "alpine 3.10 is past its support")
	}
	for _, finding := range vulns.Findings {
		assert.NotEmpty(t, finding.Fixed, "the fixable only by default: %s", finding.VulnID)
		assert.Nil(t, finding.Ignored, "none set aside by default: %s", finding.VulnID)
	}

	// An ignore of the project or of every project may set it aside: read
	// those too.
	vulns = vulnsAnswer{}
	assert.NoError(t, json.Unmarshal([]byte(f.read("get_app_vulnerabilities", merge(inApp, map[string]any{
		"severity": []string{"critical"}, "includeIgnored": true}))), &vulns))
	found := false
	for _, finding := range vulns.Findings {
		assert.Equal(t, "CRITICAL", string(finding.Severity))
		if finding.VulnID != "CVE-2021-36159" {
			continue
		}
		found = true
		assert.Equal(t, "apk-tools", finding.Package)
		assert.Equal(t, "2.10.6-r0", finding.Installed)
		assert.Equal(t, "2.10.7-r0", finding.Fixed)
		if finding.Ignored != nil {
			t.Logf("set aside by an ignore of the %q scope: %s", finding.Ignored.Scope, finding.Ignored.Reason)
		}
	}
	assert.True(t, found, "CVE-2021-36159 among %d findings", len(vulns.Findings))
}
