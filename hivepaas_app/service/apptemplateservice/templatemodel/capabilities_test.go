package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func templateWithCapabilities(t *testing.T, block string) *Template {
	t.Helper()
	tmpl := mustDecode(t)
	capabilities := map[string]any{}
	assert.NoError(t, decodeYAMLStrict([]byte(block), &capabilities))
	deployment, _ := tmpl.App["deployment"].(map[string]any)
	assert.NotNil(t, deployment, "the fixture has a deployment to hang resources on")
	deployment["resources"] = map[string]any{"capabilities": capabilities}
	return tmpl
}

func TestATemplateAsksForNoCapabilitiesUnlessItSaysSo(t *testing.T) {
	tmpl := mustDecode(t)
	capabilities, err := tmpl.Capabilities()
	assert.NoError(t, err)
	assert.Nil(t, capabilities)
	assert.False(t, tmpl.RequiresCapabilities())
}

func TestCapabilitiesAreReadFromTheTemplate(t *testing.T) {
	tmpl := templateWithCapabilities(t, `
capabilityAdd: [IPC_LOCK]
sysctls: {vm.max_map_count: "262144"}
ulimits: [{name: memlock, soft: -1, hard: -1}]
`)

	capabilities, err := tmpl.Capabilities()

	assert.NoError(t, err)
	assert.Equal(t, []string{"IPC_LOCK"}, capabilities.CapabilityAdd)
	assert.Equal(t, "262144", capabilities.Sysctls["vm.max_map_count"])
	assert.Equal(t, "memlock", capabilities.Ulimits[0].Name)
	assert.True(t, tmpl.RequiresCapabilities())
	assert.NoError(t, tmpl.Validate("demo"))
}

func TestValidateRefusesCapabilitiesItCannotShowBeforeDeploying(t *testing.T) {
	cases := map[string]struct {
		mutate func(t *testing.T) *Template
		want   string
	}{
		"a placeholder": {
			func(t *testing.T) *Template {
				return templateWithCapabilities(t, "capabilityAdd: [\"${{ params.username }}\"]\n")
			},
			"a placeholder here would make what is granted depend on what a person fills in",
		},
		"a capability that is not one": {
			func(t *testing.T) *Template { return templateWithCapabilities(t, "capabilityAdd: [ALL]\n") },
			"is not a capability such as NET_ADMIN",
		},
		"a field the block does not have": {
			func(t *testing.T) *Template { return templateWithCapabilities(t, "privileged: true\n") },
			"field privileged not found",
		},
		"a version that overrides them": {
			func(t *testing.T) *Template {
				tmpl := templateWithCapabilities(t, "capabilityAdd: [IPC_LOCK]\n")
				tmpl.Versions[0].Override = &Override{App: map[string]any{
					"deployment": map[string]any{
						"resources": map[string]any{"capabilities": map[string]any{
							"capabilityAdd": []any{"NET_ADMIN"},
						}},
					},
				}}
				return tmpl
			},
			"capabilities are declared once, in app",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.mutate(t).Validate("demo")
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
			var hpErr hperrors.HPError
			assert.ErrorAs(t, err, &hpErr)
			assert.Contains(t, hpErr.Build("en").Detail, tc.want)
		})
	}
}
