package specmodel

import (
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
		SecretsMode:       SecretsModeNone,
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
secretsMode: none
files:
    - global.yaml
    - projects/project_a/project.yaml
`, string(out))
}

func TestSecretsModeIsValid(t *testing.T) {
	assert.True(t, SecretsModeNone.IsValid())
	assert.True(t, SecretsModeEncrypted.IsValid())
	assert.True(t, SecretsModePlaintext.IsValid())
	assert.False(t, SecretsMode("").IsValid())
	assert.False(t, SecretsMode("clear").IsValid())
}

func TestSecretsModeRevealsSecrets(t *testing.T) {
	assert.False(t, SecretsModeNone.RevealsSecrets(),
		"none decrypts nothing, so it needs no capability")
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
