package specmodel

// Severity says what an issue does to the object it belongs to.
type Severity string

const (
	// SeverityFixable clears the offending part and keeps going.
	SeverityFixable Severity = "fixable"
	// SeveritySkipped drops one object and keeps the rest.
	SeveritySkipped Severity = "skipped"
	// SeverityBlocked stops everything.
	SeverityBlocked Severity = "blocked"
	// SeverityWarning clears nothing, but the result may not be what the operator
	// expects, so applying an import still needs their acceptance. Declared with
	// the others: swag lists an enum's values in the order it reads the files
	// declaring them, and that order changes between runs.
	SeverityWarning Severity = "warning"
)

// Issue codes. Export emits the first four; the rest belong to the import
// contract, and are declared here so both halves name the same things.
const (
	CodeTypeUnclassified   = "TYPE_UNCLASSIFIED"
	CodeTypeSkipped        = "TYPE_SKIPPED"
	CodePreviewAppSkipped  = "PREVIEW_APP_SKIPPED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeRefNotFound        = "REF_NOT_FOUND"
	CodeRefNotSelected     = "REF_NOT_SELECTED"
	CodeMountTargetDup     = "MOUNT_TARGET_DUPLICATED"
)

// Issue is one thing that did not go cleanly, written to be both read by a
// person and acted on by a UI.
// The json tags are not decoration. A report reaches a caller two ways: inside
// a bundle as YAML, and over the wire as the X-HivePaaS-Spec-Report header,
// which is JSON. With yaml tags alone the header came out with Go field names
// and empty strings for every unset field - "Issues", "AvailableIn": "" - which
// matches nothing else this API returns.
type Issue struct {
	Severity Severity       `yaml:"severity"          json:"severity"`
	Code     string         `yaml:"code"              json:"code"`
	Path     string         `yaml:"path"              json:"path"`
	Detail   map[string]any `yaml:"detail,omitempty"  json:"detail,omitempty"`
	// AvailableIn names the bundle file holding what the issue could not reach,
	// which is what separates "go and create this" from "select one more file".
	AvailableIn string `yaml:"availableIn,omitempty" json:"availableIn,omitempty"`
	Action      string `yaml:"action,omitempty"      json:"action,omitempty"`
	Hint        string `yaml:"hint,omitempty"        json:"hint,omitempty"`
}

// Report collects issues for one export or import run.
type Report struct {
	Issues []Issue `yaml:"issues,omitempty" json:"issues,omitempty"`
}

func (r *Report) Add(issue Issue) {
	r.Issues = append(r.Issues, issue)
}

func (r *Report) CountBySeverity(s Severity) int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == s {
			count++
		}
	}
	return count
}

func (r *Report) HasBlocked() bool {
	return r.CountBySeverity(SeverityBlocked) > 0
}

// ReportSummary is the report reduced to counts.
//
// The full report cannot travel in a response header. Measured against a
// development installation of three projects and five apps it already reached
// 6.8 KB, which is inside nginx's 8 KB default for the whole header block - a
// real installation would exceed it, and the failure would be a truncated
// header or a 502 from the proxy rather than anything legible.
//
// So the detail goes into the bundle as report.yaml, where it belongs anyway -
// somebody opening the archive months later should be able to see what was left
// out - and the header carries this instead. Its size is bounded by the number
// of declared issue codes, which is fixed.
type ReportSummary struct {
	Files  int `yaml:"files"            json:"files"`
	Issues int `yaml:"issues"           json:"issues"`
	// BySeverity and ByCode are counts, so a UI can say what happened without
	// the detail.
	BySeverity map[string]int `yaml:"bySeverity,omitempty" json:"bySeverity,omitempty"`
	ByCode     map[string]int `yaml:"byCode,omitempty"     json:"byCode,omitempty"`
	// ReportFile names where the detail is, inside the bundle.
	ReportFile string `yaml:"reportFile,omitempty" json:"reportFile,omitempty"`
}

// Summarize reduces a report to what fits in a header.
func (r *Report) Summarize(files int, reportFile string) *ReportSummary {
	summary := &ReportSummary{
		Files:      files,
		Issues:     len(r.Issues),
		BySeverity: map[string]int{},
		ByCode:     map[string]int{},
	}
	for _, issue := range r.Issues {
		summary.BySeverity[string(issue.Severity)]++
		summary.ByCode[issue.Code]++
	}
	if len(r.Issues) > 0 {
		summary.ReportFile = reportFile
	}
	return summary
}
