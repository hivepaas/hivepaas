package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestSpecExportDecisionSkipsTheThreeSkippedTypes(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeAPIKey,
		base.SettingTypeBackupSnapshot,
		base.SettingTypeApp,
	} {
		d := SpecExportDecision(&Setting{Type: typ})
		assert.False(t, d.Export, "%v must not be exported", typ)
		assert.NotEmpty(t, d.Reason, "%v must say why", typ)
	}
}

func TestSpecExportDecisionAllowsAnOrdinarySetting(t *testing.T) {
	assert.True(t, SpecExportDecision(&Setting{Type: base.SettingTypeSSLCert}).Export)
	assert.True(t, SpecExportDecision(&Setting{Type: base.SettingTypeSecret}).Export)
	assert.True(t, SpecExportDecision(&Setting{Type: base.SettingTypeAppClone}).Export)
}

// cluster-* rows created by sync live at global scope and carry a Docker object
// id; rows HivePaaS authored live at project or project-env scope and carry a
// ULID. Only the authored ones are configuration - sync rebuilds the rest, and
// it always writes ObjectScopeGlobal, so project membership is the one thing it
// can never reconstruct.
func TestSpecExportDecisionKeepsOnlyAuthoredClusterRows(t *testing.T) {
	for _, typ := range []base.SettingType{
		base.SettingTypeClusterNetwork,
		base.SettingTypeClusterVolume,
		base.SettingTypeClusterNode,
	} {
		discovered := SpecExportDecision(&Setting{Type: typ, Scope: base.ObjectScopeGlobal})
		assert.False(t, discovered.Export, "%v at global scope is sync discovery", typ)
		assert.NotEmpty(t, discovered.Reason)

		authored := SpecExportDecision(&Setting{
			Type: typ, Scope: base.ObjectScopeProject, ObjectID: "prj_1",
		})
		assert.True(t, authored.Export, "%v at project scope is configuration", typ)
	}
}

// Every setting type that can be parsed must have a policy. A type added later
// without one fails here rather than silently leaking into a spec.
func TestEverySettingTypeHasASpecPolicy(t *testing.T) {
	for typ := range settingParserMap {
		assert.NotNil(t, SpecPolicyFor(typ), "setting type %v has no registered SpecPolicy", typ)
	}
}

func TestSpecPolicyForAnUnknownTypeIsNil(t *testing.T) {
	assert.Nil(t, SpecPolicyFor(base.SettingType("not-a-real-type")))
}

func TestSpecExportDecisionRefusesAnUnregisteredType(t *testing.T) {
	d := SpecExportDecision(&Setting{Type: base.SettingType("brand-new-type")})
	assert.False(t, d.Export)
	assert.Contains(t, d.Reason, "no spec policy")
}

func TestAppRoutingStripClearsTheResetFlag(t *testing.T) {
	data := &AppRoutingSettings{Port: 8080, Reset: true}
	SpecPolicyFor(base.SettingTypeAppRouting).Strip(data)
	assert.False(t, data.Reset, "Reset is a command; importing it performs a reset")
	assert.Equal(t, 8080, data.Port)
}

func TestLoggingStripDropsManagedEndpointsOnly(t *testing.T) {
	managed := &LoggingSettings{Backend: LoggingBackend{
		Managed: true,
		Ingest:  &LoggingEndpoint{URL: "http://vl:9428"},
		Query:   &LoggingEndpoint{URL: "http://vl:9428"},
	}}
	SpecPolicyFor(base.SettingTypeLogging).Strip(managed)
	assert.Nil(t, managed.Backend.Ingest)
	assert.Nil(t, managed.Backend.Query)

	external := &LoggingSettings{Backend: LoggingBackend{
		Managed: false,
		Ingest:  &LoggingEndpoint{URL: "https://logs.example.com"},
	}}
	SpecPolicyFor(base.SettingTypeLogging).Strip(external)
	assert.NotNil(t, external.Backend.Ingest,
		"an unmanaged backend's endpoints are configuration, not derived")
}

// The values deliberately kept. This is the guard against somebody later
// "tidying up" a value that is specific to the installation but not derived.
func TestStripKeepsSystemSpecificValues(t *testing.T) {
	envVars := &EnvVars{Data: []*EnvVar{
		{Key: "HIVEPAAS_ROOT_PASSWORD", Value: "generated", IsSystem: true},
		{Key: "APP_ENV", Value: "production"},
	}}
	SpecPolicyFor(base.SettingTypeEnvVar).Strip(envVars)
	assert.Len(t, envVars.Data, 2,
		"system env vars must survive: the database volume was initialized with them")

	volume := &ClusterVolume{Managed: true, NodeID: "node-1", Driver: "local"}
	SpecPolicyFor(base.SettingTypeClusterVolume).Strip(volume)
	assert.Equal(t, "node-1", volume.NodeID, "the node id is valid on the same installation")
}
