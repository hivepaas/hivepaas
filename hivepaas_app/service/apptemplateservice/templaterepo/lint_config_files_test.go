package templaterepo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// withConfigFile adds a config file whose content is the given line to the demo
// template, which already asks for a secret parameter named password.
func withConfigFile(content string) string {
	return demoTemplateYAML + `    configFiles:
      demo.conf:
        content: |
          ` + content + `
        swarmRef: {file: {name: /etc/demo/demo.conf}}
`
}

func lintDemo(t *testing.T, demo string) string {
	t.Helper()
	_, problems := loadAndLint(t, withFile(validRepoFS(), "templates/demo.yaml", demo))
	messages := make([]string, 0, len(problems))
	for _, problem := range problems {
		messages = append(messages, problem.String())
	}
	return strings.Join(messages, "\n")
}

func TestLintAcceptsAConfigFileWithoutSecrets(t *testing.T) {
	assert.Empty(t, lintDemo(t, withConfigFile("listen = 8080")))
}

func TestLintRefusesASecretParameterInConfigFileContent(t *testing.T) {
	messages := lintDemo(t, withConfigFile(`password = "${{ params.password }}"`))

	assert.Contains(t, messages, `config file "demo.conf" writes the secret parameter "password"`)
	assert.Contains(t, messages, "declare it under settings.secrets instead")
}

func TestLintRefusesASecretParameterInAConfigFileAVersionOverrides(t *testing.T) {
	demo := strings.Replace(demoTemplateYAML,
		`  - {name: "2", release: "2.1", default: true, image: "demo:2.1.0"}`,
		`  - name: "2"
    release: "2.1"
    default: true
    image: "demo:2.1.0"
    override:
      app:
        settings:
          configFiles:
            v2.conf:
              content: 'password = "${{ params.password }}"'
              swarmRef: {file: {name: /etc/demo/v2.conf}}`, 1)

	messages := lintDemo(t, demo)

	assert.Contains(t, messages, `version 2: config file "v2.conf" writes the secret parameter "password"`)
}
