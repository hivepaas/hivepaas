package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stackTemplateYAML is an application that is several processes: a gateway with
// the domain, an authentication service behind it, and one database shared by
// both - which is the shape components exist for.
const stackTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: stack
  title: Stack
  tagline: An app that is several processes
  description: Stack.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
parameters:
  - {name: jwtSecret, title: JWT secret, type: secret, generate: {length: 32}}
dependencies:
  - {name: db, title: Database, template: demo, version: "2"}
versions:
  - name: "1"
    release: "1.0"
    default: true
    components:
      gw: {image: "gateway:1.0.0"}
      auth: {image: "auth:1.0.0"}
components:
  - name: gw
    title: Gateway
    primary: true
    needs: [auth]
    app:
      deployment:
        source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
      settings:
        kind: {category: webapp, engine: stack, webapp: {}}
        envVars:
          data:
            - {k: AUTH_URL, v: "http://${{ comp.auth.key }}:9000"}
            - {k: JWT_SECRET, v: "${{ params.jwtSecret }}"}
        routing: {port: 8000, exposePublicly: true}
  - name: auth
    title: Auth
    app:
      deployment:
        source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
      settings:
        kind: {category: webapp, engine: stack, webapp: {}}
        envVars:
          data:
            - {k: JWT_SECRET, v: "${{ params.jwtSecret }}"}
            - {k: DB_HOST, v: "${{ deps.db.ref.HIVEPAAS_HOST }}"}
            - {k: DB_PASSWORD, v: "${{ deps.db.ref.HIVEPAAS_PASSWORD }}"}
`

func lintStack(t *testing.T, stack string) []string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/stack.yaml", stack))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return messages
}

func TestLintAcceptsATemplateWithComponents(t *testing.T) {
	assert.Empty(t, lintStack(t, stackTemplateYAML))
}

func TestLintChecksComponents(t *testing.T) {
	cases := map[string]struct {
		stack string
		want  string
	}{
		"no primary": {
			strings.Replace(stackTemplateYAML, "    primary: true\n", "", 1),
			"exactly one component must be primary",
		},
		"two primaries": {
			strings.Replace(stackTemplateYAML, "  - name: auth\n    title: Auth\n",
				"  - name: auth\n    title: Auth\n    primary: true\n", 1),
			"exactly one component must be primary",
		},
		"needs a component that is not declared": {
			strings.Replace(stackTemplateYAML, "needs: [auth]", "needs: [cache]", 1),
			`components[gw].needs: "cache" is not a declared component`,
		},
		"needs itself": {
			strings.Replace(stackTemplateYAML, "needs: [auth]", "needs: [gw]", 1),
			"components[gw].needs: a component cannot need itself",
		},
		"a version pins no image for a component": {
			strings.Replace(stackTemplateYAML, "      auth: {image: \"auth:1.0.0\"}\n", "", 1),
			"versions[1].components[auth].image is required",
		},
		"a version pins an image for something that is not a component": {
			strings.Replace(stackTemplateYAML, "      auth: {image: \"auth:1.0.0\"}",
				"      auth: {image: \"auth:1.0.0\"}\n      ghost: {image: \"ghost:1.0.0\"}", 1),
			`versions[1].components: "ghost" is not a declared component`,
		},
		"a moving tag": {
			strings.Replace(stackTemplateYAML, "auth:1.0.0", "auth:latest", 1),
			"must name an exact release",
		},
		"an app block beside the components": {
			stackTemplateYAML + "\napp:\n  deployment:\n    source: {activeMethod: image}\n",
			"components and app are exclusive",
		},
		"a domain on a component that is not the primary": {
			strings.Replace(stackTemplateYAML,
				"            - {k: DB_PASSWORD, v: \"${{ deps.db.ref.HIVEPAAS_PASSWORD }}\"}",
				"            - {k: DB_PASSWORD, v: \"${{ deps.db.ref.HIVEPAAS_PASSWORD }}\"}\n"+
					"        routing: {port: 9000, exposePublicly: true}", 1),
			"only the primary component may be exposed publicly",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			messages := lintStack(t, tc.stack)
			assert.NotEmpty(t, messages)
			assert.Contains(t, strings.Join(messages, "; "), tc.want)
		})
	}
}

// A component reading what another shares only works if that one is rendered
// first, and needs is what says so. The linter renders in that order, so it is
// the linter that catches the missing need rather than the person installing it.
func TestLintCatchesAComponentReadingOneRenderedLater(t *testing.T) {
	stack := strings.Replace(stackTemplateYAML, "needs: [auth]", "needs: []", 1)
	stack = strings.Replace(stack, `- {k: AUTH_URL, v: "http://${{ comp.auth.key }}:9000"}`,
		`- {k: AUTH_HOST, v: "${{ comp.auth.ref.HIVEPAAS_HOST }}"}`, 1)

	messages := lintStack(t, stack)

	assert.Contains(t, strings.Join(messages, "; "), "which is rendered later: name it in needs")
}

// The key is what an app answers to on the network, and two components whose
// names truncate to the same one would collide after they were created.
func TestLintAcceptsComponentKeysThatStayDistinct(t *testing.T) {
	assert.Empty(t, lintStack(t, stackTemplateYAML))
}
