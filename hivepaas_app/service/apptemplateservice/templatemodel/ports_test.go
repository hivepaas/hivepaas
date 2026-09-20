package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func templateWithPorts(t *testing.T, ports string) *Template {
	t.Helper()
	tmpl := mustDecode(t)
	entries := []any{}
	assert.NoError(t, yaml.Unmarshal([]byte(ports), &entries))
	deployment, _ := tmpl.App["deployment"].(map[string]any)
	assert.NotNil(t, deployment)
	deployment["networks"] = map[string]any{"endpointSpec": map[string]any{"ports": entries}}
	return tmpl
}

func TestATemplatePublishesNoPortsUnlessItSaysSo(t *testing.T) {
	assert.Empty(t, mustDecode(t).PublishedPorts())
}

func TestPublishedPortsAreReadFromTheTemplate(t *testing.T) {
	tmpl := templateWithPorts(t, `
- {target: 53, published: 53, protocol: udp, publishMode: host}
- {target: 8080, published: 8080}
`)

	ports := tmpl.PublishedPorts()

	assert.Equal(t, []PublishedPort{
		{Target: 53, Published: 53, Protocol: "udp", PublishMode: "host"},
		{Target: 8080, Published: 8080},
	}, ports)
}

// A port a parameter chooses is reported with that parameter's default and the
// parameter's name, so that what is shown before deploying can follow what a
// person types instead of the default they are about to change.
func TestAPortAParameterChoosesNamesTheParameter(t *testing.T) {
	tmpl := templateWithPorts(t, `
- {target: "${{ params.vpnPort }}", published: "${{ params.vpnPort }}", protocol: udp}
`)
	tmpl.Parameters = append(tmpl.Parameters, &Parameter{
		Name: "vpnPort", Title: "VPN port", Type: ParamTypeInt, Default: 51820,
	})

	ports := tmpl.PublishedPorts()

	assert.Equal(t, []PublishedPort{
		{Target: 51820, Published: 51820, PublishedParam: "vpnPort", Protocol: "udp"},
	}, ports)
}

func TestAPortThatCannotBeReadIsReportedAsUnknown(t *testing.T) {
	tmpl := templateWithPorts(t, `
- {target: "${{ params.missing }}", published: "${{ params.missing }}"}
- {target: 70000, published: 70000}
`)

	assert.Equal(t, []PublishedPort{
		{PublishedParam: "missing"},
		{},
	}, tmpl.PublishedPorts())
}
