package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// imageScanRoutes answers api's image scan: two images, a finding of each
// kind the tool sorts out, and one an ignore sets aside.
type imageScanRoutes struct {
	off bool
	// extra is how many more fixable HIGH findings the scan has.
	extra int
}

func (r *imageScanRoutes) add(api *gin.RouterGroup) {
	finding := func(image, id, severity, pkg, fixed string) gin.H {
		return gin.H{"imageId": image, "vulnId": id, "severity": severity, "pkgName": pkg, "pkgType": "alpine",
			"installedVersion": "1.0", "fixedVersion": fixed, "title": "a flaw in " + pkg,
			"url": "https://avd.test/" + id}
	}
	api.GET("/projects/p1/prod/apps/a1/image-scan", func(ctx *gin.Context) {
		findings := []gin.H{
			finding("sha256:aa", "CVE-1", "CRITICAL", "openssl", "3.0.9"),
			finding("sha256:bb", "CVE-2", "HIGH", "busybox", ""),
			finding("sha256:aa", "CVE-3", "MEDIUM", "zlib", "1.3"),
		}
		for i := range r.extra {
			findings = append(findings, finding("sha256:aa", fmt.Sprintf("CVE-9%04d", i), "HIGH",
				strings.Repeat("p", 200), "2.0"))
		}
		ignored := finding("sha256:bb", "CVE-4", "HIGH", "musl", "1.2.5")
		ignored["ignore"] = gin.H{"id": "i1", "scope": "project", "reason": "never fed outside input",
			"setBy": "ada", "inheritable": true}
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"enabled": !r.off, "scannedAt": "2026-10-15T02:00:00Z",
			"summary": gin.H{"critical": gin.H{"total": 1, "fixable": 1}, "high": gin.H{"total": 1}, "ignored": 1},
			"images": []gin.H{
				{"imageId": "sha256:aa", "imageRef": "shop/api:2", "status": "done", "osEosl": false},
				{"imageId": "sha256:bb", "imageRef": "shop/sidecar:1", "status": "done", "osEosl": true},
			},
			"findings": findings, "ignored": []gin.H{ignored},
		}})
	})
	api.POST("/projects/p1/prod/apps/a1/image-scan", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"task": gin.H{"id": "t7"}}})
	})
}

func vulnsSession(t *testing.T, routes *imageScanRoutes) func(args map[string]any) (vulnsAnswer, string, bool) {
	t.Helper()
	w := newMCPWorld(t, (&appRoutes{}).add, routes.add)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return func(args map[string]any) (vulnsAnswer, string, bool) {
		text, isErr := callTool(t, session, "get_app_vulnerabilities", args)
		var out vulnsAnswer
		if !isErr {
			assert.NoError(t, json.Unmarshal([]byte(text), &out))
		}
		return out, text, isErr
	}
}

func vulnIDs(findings []*vulnFinding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.VulnID)
	}
	return out
}

// By default an app's vulnerabilities are what upgrading removes: CRITICAL
// and HIGH with a fix out, none an ignore sets aside.
func TestAppVulnerabilitiesAreTheFixableWorstByDefault(t *testing.T) {
	call := vulnsSession(t, &imageScanRoutes{})
	out, text, isErr := call(shopProd("api"))
	assert.False(t, isErr, text)
	assert.Equal(t, []string{"CVE-1"}, vulnIDs(out.Findings))
	assert.Equal(t, 1, out.Matched)
	assert.Zero(t, out.Omitted)
	assert.Equal(t, "shop/api:2", out.Findings[0].Image, "an app of two images says which")
	assert.Equal(t, "3.0.9", out.Findings[0].Fixed)
	assert.True(t, out.Enabled)
	assert.Len(t, out.Images, 2)
	assert.Equal(t, 1, out.Summary.Critical.Fixable)
	assert.NotContains(t, text, "avd.test", "a finding's link is not what a model reads")
}

func TestAppVulnerabilitiesWidened(t *testing.T) {
	call := vulnsSession(t, &imageScanRoutes{})
	args := shopProd("api")
	args["severity"] = []any{"high", "Medium"}
	args["includeUnfixed"] = true
	args["includeIgnored"] = true
	out, text, isErr := call(args)
	assert.False(t, isErr, text)
	assert.Equal(t, []string{"CVE-2", "CVE-3", "CVE-4"}, vulnIDs(out.Findings), "the ignored ones last")
	if ignored := out.Findings[2].Ignored; assert.NotNil(t, ignored) {
		assert.Equal(t, "never fed outside input", ignored.Reason)
		assert.Equal(t, "ada", ignored.SetBy)
	}
	assert.Nil(t, out.Findings[0].Ignored)
}

func TestAppVulnerabilitiesRefuseAnUnknownSeverity(t *testing.T) {
	call := vulnsSession(t, &imageScanRoutes{})
	args := shopProd("api")
	args["severity"] = []any{"severe"}
	_, text, isErr := call(args)
	assert.True(t, isErr)
	assert.Contains(t, text, `"severe"`)
}

// Findings past what an answer holds are counted, not answered.
func TestAppVulnerabilitiesFitAnAnswer(t *testing.T) {
	call := vulnsSession(t, &imageScanRoutes{extra: 1000})
	out, text, isErr := call(shopProd("api"))
	assert.False(t, isErr, text)
	assert.LessOrEqual(t, len(text), MaxToolOutput)
	assert.Equal(t, 1001, out.Matched)
	assert.Positive(t, out.Omitted)
	assert.Equal(t, out.Matched, len(out.Findings)+out.Omitted)
	assert.Equal(t, "CVE-1", out.Findings[0].VulnID, "the worst kept")
}

func TestAppVulnerabilitiesSayScanningIsOff(t *testing.T) {
	call := vulnsSession(t, &imageScanRoutes{off: true})
	out, text, isErr := call(shopProd("api"))
	assert.False(t, isErr, text)
	assert.False(t, out.Enabled)
	assert.Equal(t, noteScanOff, out.Note)
}

func newImageScanWorld(t *testing.T, scans *imageScanRoutes) *writeWorld {
	t.Helper()
	routes := &writeRoutes{}
	w := newMCPWorld(t, routes.add, (&appRoutes{}).add, scans.add)
	w.sw.allowWrite = true
	return &writeWorld{mcpWorld: w, routes: routes}
}

func TestScanningAnAppsImages(t *testing.T) {
	w := newImageScanWorld(t, &imageScanRoutes{})
	session := w.session(t, "key1")
	text, isErr := callTool(t, session, "plan_scan_app_images", shopProd("api"))
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"critical"`, "the plan shows the last scan's counts")
	assert.Empty(t, w.routes.writes(), "a plan scans nothing")

	text, isErr = callTool(t, session, "apply_plan", map[string]any{"planToken": planOf(t, text).PlanToken})
	assert.False(t, isErr, text)
	assert.Contains(t, text, "t7")
	if sent := w.routes.writes(); assert.Len(t, sent, 1) {
		assert.Equal(t, "POST", sent[0].Method)
		assert.Equal(t, "/projects/p1/prod/apps/a1/image-scan", sent[0].Path)
	}
}

func TestNoScanIsPlannedWhileScanningIsOff(t *testing.T) {
	w := newImageScanWorld(t, &imageScanRoutes{off: true})
	text, isErr := callTool(t, w.session(t, "key1"), "plan_scan_app_images", shopProd("api"))
	assert.False(t, isErr, text)
	plan := planOf(t, text)
	assert.Empty(t, plan.PlanToken)
	assert.Contains(t, string(plan.Plan), "switched off")
}
