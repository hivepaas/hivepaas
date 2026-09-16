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
