package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// The tools of an app's image scan: what its images carry, and a scan now.

const (
	noteScanOff = "image scanning is switched off: System > Image Scanning switches it on; what is listed " +
		"is what the last scans found"
	noteNeverScanned = "the app's images were never scanned; plan_scan_app_images scans them"
)

// ---- get_app_vulnerabilities ----

type appVulnsInput struct {
	Project        string   `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env            string   `json:"env" jsonschema:"the env's name, such as production"`
	App            string   `json:"app" jsonschema:"the app's key, name or id; list_apps lists them"`
	Severity       []string `json:"severity,omitempty" jsonschema:"the severities whose findings to list: critical, high, medium, low, unknown; critical and high when not given"` //nolint:lll
	IncludeUnfixed bool     `json:"includeUnfixed,omitempty" jsonschema:"true to list the findings no fixed version is out for too; only those with a fix when not given"`         //nolint:lll
	IncludeIgnored bool     `json:"includeIgnored,omitempty" jsonschema:"true to list the findings an ignore sets aside too, after the others, each with the ignore's reason"`     //nolint:lll
}

// severities are the severities asked for, as the API spells them.
func (in appVulnsInput) severities() ([]base.VulnSeverity, error) {
	if len(in.Severity) == 0 {
		return []base.VulnSeverity{base.VulnSeverityCritical, base.VulnSeverityHigh}, nil
	}
	out := make([]base.VulnSeverity, 0, len(in.Severity))
	for _, s := range in.Severity {
		severity := base.VulnSeverity(strings.ToUpper(strings.TrimSpace(s)))
		if !slices.Contains(base.AllVulnSeverities, severity) {
			return nil, &InputError{Message: fmt.Sprintf("severity %q is none of critical, high, medium, low "+
				"and unknown", s)}
		}
		out = append(out, severity)
	}
	return out, nil
}

// vulnFinding is a finding as a model reads it: what to upgrade, from what
// to what.
type vulnFinding struct {
	VulnID    string            `json:"vulnId"`
	Severity  base.VulnSeverity `json:"severity"`
	Package   string            `json:"package"`
	PkgPath   string            `json:"pkgPath,omitempty"`
	Installed string            `json:"installed"`
	Fixed     string            `json:"fixed,omitempty"`
	Title     string            `json:"title,omitempty"`
	// Image is the image's reference, when the app runs more than one.
	Image string `json:"image,omitempty"`
	// Ignored is the ignore that sets it aside.
	Ignored *vulnIgnore `json:"ignored,omitempty"`
}

type vulnIgnore struct {
	Scope    base.ObjectScopeType `json:"scope"`
	Reason   string               `json:"reason"`
	SetBy    string               `json:"setBy,omitempty"`
	ExpireAt *time.Time           `json:"expireAt,omitempty"`
}

type vulnsAnswer struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	// Enabled says image scanning is switched on.
	Enabled   bool       `json:"enabled"`
	ScannedAt *time.Time `json:"scannedAt,omitempty"`
	Error     string     `json:"error,omitempty"`
	// Scanning is the app's scan that has not ended yet.
	Scanning *appimagescandto.ScanTaskResp `json:"scanning,omitempty"`
	// Summary counts every finding of every image, after ignores.
	Summary *appimagescandto.SummaryResp `json:"summary"`
	Images  []*appimagescandto.ImageResp `json:"images"`
	// Findings are those asked for, the worst first.
	Findings []*vulnFinding `json:"findings"`
	// Matched is how many findings were asked for; Omitted, how many of them
	// were left out to keep the answer under 64 KB.
	Matched int    `json:"matched"`
	Omitted int    `json:"omitted,omitempty"`
	Note    string `json:"note,omitempty"`
}

func getAppVulnerabilitiesTool() Tool {
	return readTool("get_app_vulnerabilities", "Get an app's vulnerabilities",
		"GET /projects/{project}/{env}/apps/{app}/image-scan, cut to what a model needs: the known "+
			"vulnerabilities of the packages in the images the app runs, as Trivy found them when it last "+
			"scanned them - the app's Security tab. The answer has each image with its counts by severity, "+
			"the counts of every image (summary), and the findings asked for, the worst first: each "+
			"vulnerability's ID, its package, the version installed and the one that fixes it. By default "+
			"those CRITICAL and HIGH with a fix out, which upgrading the package or the base image removes. "+
			"An image past its OS's support (osEosl) gets no fixes at all: its base image is what to "+
			"change. Findings someone set aside with a reason are left out unless includeIgnored. "+
			"plan_scan_app_images scans again.",
		func(ctx context.Context, call *Call, in appVulnsInput) (vulnsAnswer, error) {
			severities, err := in.severities()
			if err != nil {
				return vulnsAnswer{}, err
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return vulnsAnswer{}, err
			}
			var resp appimagescandto.GetAppImageScanResp
			if err = call.Get(ctx, ref.path("/image-scan"), nil, &resp); err != nil {
				return vulnsAnswer{}, err
			}
			out := vulnsAnswer{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Findings: []*vulnFinding{}}
			if resp.Data == nil {
				return out, nil
			}
			return vulnsOf(out, resp.Data, severities, in.IncludeUnfixed, in.IncludeIgnored)
		})
}

// vulnsOf fills an answer with a scan's findings of the severities asked for,
// as many as fit.
func vulnsOf(out vulnsAnswer, scan *appimagescandto.AppImageScanResp, severities []base.VulnSeverity,
	unfixed, ignored bool) (vulnsAnswer, error) {
	out.Enabled, out.ScannedAt, out.Error = scan.Enabled, scan.ScannedAt, scan.Error
	out.Scanning, out.Summary, out.Images = scan.Task, scan.Summary, scan.Images
	switch {
	case !scan.Enabled:
		out.Note = noteScanOff
	case scan.ScannedAt == nil && scan.Task == nil:
		out.Note = noteNeverScanned
	}

	refs := map[string]string{}
	if len(scan.Images) > 1 {
		for _, image := range scan.Images {
			refs[image.ImageID] = image.ImageRef
		}
	}
	wanted := func(f *appimagescandto.FindingResp) bool {
		return f != nil && slices.Contains(severities, f.Severity) && (unfixed || f.FixedVersion != "")
	}
	var matched []*vulnFinding
	for _, f := range scan.Findings {
		if wanted(f) {
			matched = append(matched, toVulnFinding(f, refs[f.ImageID]))
		}
	}
	if ignored {
		for _, f := range scan.Ignored {
			if f == nil || !wanted(f.FindingResp) {
				continue
			}
			finding := toVulnFinding(f.FindingResp, refs[f.ImageID])
			if f.Ignore != nil {
				finding.Ignored = &vulnIgnore{Scope: f.Ignore.Scope, Reason: f.Ignore.Reason, SetBy: f.Ignore.SetBy,
					ExpireAt: f.Ignore.ExpireAt}
			}
			matched = append(matched, finding)
		}
	}
	out.Matched = len(matched)

	// The rest of the answer first, then the findings while they fit: the
	// Omitted a cut adds is a few bytes, which the margin holds.
	const margin = 64
	raw, err := json.Marshal(out)
	if err != nil {
		return vulnsAnswer{}, fmt.Errorf("mcp: encoding the answer: %w", err)
	}
	size := len(raw) + margin
	for _, f := range matched {
		raw, err = json.Marshal(f)
		if err != nil {
			return vulnsAnswer{}, fmt.Errorf("mcp: encoding the answer: %w", err)
		}
		if size+len(raw)+1 > MaxToolOutput {
			break
		}
		size += len(raw) + 1
		out.Findings = append(out.Findings, f)
	}
	out.Omitted = len(matched) - len(out.Findings)
	return out, nil
}

func toVulnFinding(f *appimagescandto.FindingResp, image string) *vulnFinding {
	return &vulnFinding{VulnID: f.VulnID, Severity: f.Severity, Package: f.PkgName, PkgPath: f.PkgPath,
		Installed: f.InstalledVersion, Fixed: f.FixedVersion, Title: f.Title, Image: image}
}

// ---- plan_scan_app_images ----

type scanAppPlan struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	// ScannedAt and Summary are the last scan's, which the new one replaces.
	ScannedAt *time.Time                   `json:"scannedAt,omitempty"`
	Summary   *appimagescandto.SummaryResp `json:"summary,omitempty"`
	// Scanning is the app's scan that has not ended yet.
	Scanning *appimagescandto.ScanTaskResp `json:"scanning,omitempty"`
	Effect   string                        `json:"effect"`
	Issues   []string                      `json:"issues,omitempty"`
}

func planScanAppImagesTool() Tool {
	return planTool("plan_scan_app_images", "Plan scanning an app's images",
		"Plans POST /projects/{project}/{env}/apps/{app}/image-scan: Trivy scans the images the app's "+
			"containers run now for known vulnerabilities, as the Security tab's Scan now does - after a "+
			"deployment, or to see whether a new advisory touches it. The app is not touched. The plan shows "+
			"the last scan's counts. Applied, the endpoint answers the task that scans. Nothing runs until "+
			"apply_plan.",
		NeedWrite, &applier{follow: "get_task with the task's id shows how the scan goes; " +
			"get_app_vulnerabilities what it found, once it has ended."},
		func(ctx context.Context, call *Call, in appInput) (scanAppPlan, *storedPlan, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return scanAppPlan{}, nil, err
			}
			var resp appimagescandto.GetAppImageScanResp
			if err = call.Get(ctx, ref.path("/image-scan"), nil, &resp); err != nil {
				return scanAppPlan{}, nil, err
			}
			out := scanAppPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env}
			if resp.Data == nil || !resp.Data.Enabled {
				out.Issues = []string{"image scanning is switched off: System > Image Scanning switches it on"}
				return out, nil, nil
			}
			out.ScannedAt, out.Summary, out.Scanning = resp.Data.ScannedAt, resp.Data.Summary, resp.Data.Task
			out.Effect = "Trivy scans each image the app runs, as a task; what it finds replaces the last scan's"
			if out.Scanning != nil && out.Scanning.Status == base.TaskStatusNotStarted {
				out.Effect += ". A scan of the app is queued and has not started: the endpoint answers that one"
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/image-scan"),
				Body:    json.RawMessage(`{}`),
				Summary: fmt.Sprintf("scan the images of %s in %s/%s", ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}
