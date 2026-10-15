package appimagescandto

import (
	"cmp"
	"slices"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// CountResp is how many findings of a severity, and how many of them have a
// fixed version.
type CountResp struct {
	Total   int `json:"total"`
	Fixable int `json:"fixable"`
}

// SummaryResp counts findings by severity, after ignores.
type SummaryResp struct {
	Critical CountResp `json:"critical"`
	High     CountResp `json:"high"`
	Medium   CountResp `json:"medium"`
	Low      CountResp `json:"low"`
	Unknown  CountResp `json:"unknown"`
	// Ignored is how many were set aside, of every severity.
	Ignored int `json:"ignored"`
}

func TransformSummary(s imagescanservice.Summary) *SummaryResp {
	count := func(c imagescanservice.Count) CountResp { return CountResp{Total: c.Total, Fixable: c.Fixable} }
	return &SummaryResp{Critical: count(s.Critical), High: count(s.High), Medium: count(s.Medium),
		Low: count(s.Low), Unknown: count(s.Unknown), Ignored: s.Ignored}
}

// ImageResp is an image an app runs, and its last scan.
type ImageResp struct {
	ImageID  string `json:"imageId"`
	ImageRef string `json:"imageRef"`
	Platform string `json:"platform,omitempty"`
	OSFamily string `json:"osFamily,omitempty"`
	OSName   string `json:"osName,omitempty"`
	// OSEOSL says the OS is past its support: its packages get no fixes.
	OSEOSL bool                 `json:"osEosl"`
	Status base.ImageScanStatus `json:"status"`
	// Source is where the last scan read the image: on the node running it, or
	// from its registry when scans run on another node.
	Source base.ImageScanSource `json:"source,omitempty"`
	// Error is why the last scan failed; the findings are the scan's before.
	Error        string       `json:"error,omitempty"`
	ScannedAt    time.Time    `json:"scannedAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DBUpdatedAt  *time.Time   `json:"dbUpdatedAt,omitempty"`
	TrivyVersion string       `json:"trivyVersion,omitempty"`
	Summary      *SummaryResp `json:"summary"`
}

// FindingResp is one vulnerability of one package of an image.
type FindingResp struct {
	ImageID          string            `json:"imageId"`
	VulnID           string            `json:"vulnId"`
	PkgName          string            `json:"pkgName"`
	PkgType          string            `json:"pkgType"`
	PkgPath          string            `json:"pkgPath,omitempty"`
	InstalledVersion string            `json:"installedVersion"`
	FixedVersion     string            `json:"fixedVersion,omitempty"`
	Severity         base.VulnSeverity `json:"severity"`
	Title            string            `json:"title,omitempty"`
	URL              string            `json:"url,omitempty"`
	PublishedAt      *time.Time        `json:"publishedAt,omitempty"`
}

// IgnoredFindingResp is a finding an ignore sets aside, and the ignore.
type IgnoredFindingResp struct {
	*FindingResp
	Ignore *IgnoreRefResp `json:"ignore"`
}

// IgnoreRefResp is the ignore that sets a finding aside: enough to say why,
// and to find it at its scope.
type IgnoreRefResp struct {
	ID       string               `json:"id"`
	Scope    base.ObjectScopeType `json:"scope"`
	Package  string               `json:"package,omitempty"`
	Reason   string               `json:"reason"`
	SetBy    string               `json:"setBy,omitempty"`
	ExpireAt *time.Time           `json:"expireAt,omitempty"`
	// Inheritable says the apps of the scope's children have it too; a global
	// one that is not applies to HivePaaS's own parts alone.
	Inheritable bool `json:"inheritable"`
}

func TransformImage(scan *entity.ImageScan, summary imagescanservice.Summary) *ImageResp {
	resp := &ImageResp{
		ImageID:      scan.ImageID,
		ImageRef:     scan.ImageRef,
		Platform:     scan.Platform,
		OSFamily:     scan.OSFamily,
		OSName:       scan.OSName,
		OSEOSL:       scan.OSEOSL,
		Status:       scan.Status,
		Source:       scan.Source,
		Error:        scan.Error,
		ScannedAt:    scan.ScannedAt,
		UpdatedAt:    scan.UpdatedAt,
		TrivyVersion: scan.TrivyVersion,
		Summary:      TransformSummary(summary),
	}
	if !scan.DBUpdatedAt.IsZero() {
		resp.DBUpdatedAt = &scan.DBUpdatedAt
	}
	return resp
}

func TransformFinding(f *entity.ImageScanFinding) *FindingResp {
	resp := &FindingResp{
		ImageID:          f.ImageID,
		VulnID:           f.VulnID,
		PkgName:          f.PkgName,
		PkgType:          f.PkgType,
		PkgPath:          f.PkgPath,
		InstalledVersion: f.InstalledVersion,
		FixedVersion:     f.FixedVersion,
		Severity:         f.Severity,
		Title:            f.Title,
		URL:              f.URL,
	}
	if !f.PublishedAt.IsZero() {
		resp.PublishedAt = &f.PublishedAt
	}
	return resp
}

func TransformIgnoreRef(ignore *imagescanservice.Ignore) *IgnoreRefResp {
	resp := &IgnoreRefResp{
		ID:      ignore.Setting.ID,
		Scope:   ignore.Setting.Scope,
		Package: ignore.Data.Package,
		Reason:  ignore.Data.Reason,
		SetBy:   ignore.Data.SetBy,

		Inheritable: ignore.Setting.Inheritable,
	}
	if !ignore.Setting.ExpireAt.IsZero() {
		resp.ExpireAt = &ignore.Setting.ExpireAt
	}
	return resp
}

// SortFindings puts the worst first: by severity, the fixable before the rest,
// then by package and vulnerability.
func SortFindings(findings []*entity.ImageScanFinding) {
	slices.SortStableFunc(findings, func(a, b *entity.ImageScanFinding) int {
		return cmp.Or(
			cmp.Compare(a.Severity.Rank(), b.Severity.Rank()),
			-cmp.Compare(boolInt(a.Fixable()), boolInt(b.Fixable())),
			cmp.Compare(a.PkgName, b.PkgName),
			cmp.Compare(a.VulnID, b.VulnID),
		)
	})
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ScannedAppResp is an app and what its images carry, in counts: a row of
// System > Image Scanning > Apps and of a project's Security tab, which list
// the worst first, by fixable CRITICAL, then fixable HIGH.
type ScannedAppResp struct {
	App        *basedto.NamedObjectResp `json:"app"`
	Project    *basedto.NamedObjectResp `json:"project"`
	ProjectEnv *EnvRefResp              `json:"projectEnv"`
	ScannedAt  time.Time                `json:"scannedAt"`
	Error      string                   `json:"error,omitempty"`
	Summary    *SummaryResp             `json:"summary"`
	Images     []*ImageResp             `json:"images"`
}

// EnvRefResp is an app's env: its key is what an app's address names.
type EnvRefResp struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}
