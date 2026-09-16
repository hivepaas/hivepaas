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
