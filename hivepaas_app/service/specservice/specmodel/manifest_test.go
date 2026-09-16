package specmodel

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestManifestMarshalsWithStableFieldOrder(t *testing.T) {
	m := &Manifest{
		APIVersion:        APIVersion,
		Kind:              KindSpec,
		ExportedAt:        time.Date(2026, 9, 16, 10, 15, 0, 0, time.UTC),
		SourceAppVersion:  "v0.1.0",
		SourceVersionCode: "v000001",
		Scope:             "global",
		SecretsMode:       SecretsModeOmit,
		Files:             []string{"global.yaml", "projects/project_a/project.yaml"},
	}

	out, err := yaml.Marshal(m)
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: hivepaas.com/v1
kind: Spec
exportedAt: 2026-09-16T10:15:00Z
sourceAppVersion: v0.1.0
sourceVersionCode: v000001
scope: global
secretsMode: omit
files:
    - global.yaml
    - projects/project_a/project.yaml
`, string(out))
}

func TestSecretsModeIsValid(t *testing.T) {
	assert.True(t, SecretsModeOmit.IsValid())
	assert.True(t, SecretsModeEncrypted.IsValid())
	assert.True(t, SecretsModePlaintext.IsValid())
	assert.False(t, SecretsMode("").IsValid())
	assert.False(t, SecretsMode("none").IsValid(), "none was renamed to omit")
	assert.False(t, SecretsMode("clear").IsValid())
}

func TestSecretsModeRevealsSecrets(t *testing.T) {
	assert.False(t, SecretsModeOmit.RevealsSecrets(),
		"omit decrypts nothing, so it needs no capability")
	assert.True(t, SecretsModeEncrypted.RevealsSecrets())
	assert.True(t, SecretsModePlaintext.RevealsSecrets())
}

func TestReportCountsBySeverity(t *testing.T) {
	r := &Report{}
	r.Add(Issue{Severity: SeverityFixable, Code: CodeRefNotFound, Path: "a"})
	r.Add(Issue{Severity: SeveritySkipped, Code: CodeTypeSkipped, Path: "b"})
	r.Add(Issue{Severity: SeverityFixable, Code: CodeRefNotFound, Path: "c"})

	assert.Len(t, r.Issues, 3)
	assert.Equal(t, 2, r.CountBySeverity(SeverityFixable))
	assert.Equal(t, 1, r.CountBySeverity(SeveritySkipped))
	assert.Equal(t, 0, r.CountBySeverity(SeverityBlocked))
	assert.False(t, r.HasBlocked())

	r.Add(Issue{Severity: SeverityBlocked, Code: "X", Path: "d"})
	assert.True(t, r.HasBlocked())
}

// A report travels as YAML inside a bundle and as JSON in a response header, so
// it needs both sets of tags. With yaml tags alone the header came out with Go
// field names and an empty string for every unset field.
func TestReportMarshalsAsCamelCaseJSON(t *testing.T) {
	out, err := json.Marshal(&Report{Issues: []Issue{{
		Severity: SeverityFixable,
		Code:     CodeRefNotSelected,
		Path:     "projects/a/envs/dev/apps/api/routing",
	}}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"issues":[{
		"severity":"fixable",
		"code":"REF_NOT_SELECTED",
		"path":"projects/a/envs/dev/apps/api/routing"
	}]}`, string(out))
}

// The full report cannot travel in a response header: on a development
// installation of three projects and five apps it already reached 6.8 KB,
// inside nginx's 8 KB default for the whole header block. The summary must stay
// small however many issues there are.
func TestReportSummaryStaysSmallAtScale(t *testing.T) {
	report := &Report{}
	for i := range 2000 {
		report.Add(Issue{
			Severity: SeveritySkipped,
			Code:     CodeTypeSkipped,
			Path:     "projects/p/envs/dev/apps/app-" + string(rune('a'+i%26)) + "/some/long/setting/path",
			Detail:   map[string]any{"type": "ssl-cert", "id": "01JAB9XED0GTXBSQDFVYAJ8WM2"},
			Action:   "a reason long enough to be useful to somebody reading it later",
		})
	}

	full, err := json.Marshal(report)
	assert.NoError(t, err)
	assert.Greater(t, len(full), 100_000, "precondition: the full report is far too big for a header")

	encoded, err := json.Marshal(report.Summarize(12, "report.yaml"))
	assert.NoError(t, err)
	assert.Less(t, len(encoded), 512,
		"the summary is bounded by the number of issue codes, not by the number of issues")
}

func TestReportSummaryCountsAndPointsAtTheFile(t *testing.T) {
	report := &Report{}
	report.Add(Issue{Severity: SeveritySkipped, Code: CodeTypeSkipped})
	report.Add(Issue{Severity: SeveritySkipped, Code: CodeTypeSkipped})
	report.Add(Issue{Severity: SeverityFixable, Code: CodeServiceUnavailable})

	summary := report.Summarize(9, "report.yaml")
	assert.Equal(t, 9, summary.Files)
	assert.Equal(t, 3, summary.Issues)
	assert.Equal(t, 2, summary.ByCode[CodeTypeSkipped])
	assert.Equal(t, 1, summary.ByCode[CodeServiceUnavailable])
	assert.Equal(t, 2, summary.BySeverity[string(SeveritySkipped)])
	assert.Equal(t, "report.yaml", summary.ReportFile)
}

// A clean export names no report file, so a UI has nothing to point at.
func TestReportSummaryNamesNoFileWhenThereAreNoIssues(t *testing.T) {
	summary := (&Report{}).Summarize(9, "report.yaml")
	assert.Equal(t, 0, summary.Issues)
	assert.Empty(t, summary.ReportFile)
}
