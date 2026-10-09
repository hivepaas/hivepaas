package envvarservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// usesHeldSecret is a variable built from a secret its app does not make
// inheritable, as its preview inherits it: emptied, with the error that says so.
func usesHeldSecret() *EnvVar {
	return &EnvVar{
		EnvVar: &entity.EnvVar{Key: "USES"},
		Errors: []*ParseError{{Type: ParseErrorVarUsesWithheldSecret, Name: "USES", Secrets: []string{"HELD"}}},
	}
}

// A preview goes without the secrets its app keeps from previews: a variable
// built from one is empty there, as the pull request's comment says, and is no
// error that stops the preview.
func TestAPreviewGoesWithoutTheSecretsItsAppKeeps(t *testing.T) {
	data := &AppEnvVarData{App: &entity.App{Name: "web-pr", ParentID: "web"}, EnvVars: []*EnvVar{usesHeldSecret()}}

	assert.Empty(t, data.Errors())
}

// An app inheriting a variable built from a secret its env or project keeps to
// itself is refused it: that is an error to fix there.
func TestAnAppGivenAVariableOfAHeldSecretIsRefused(t *testing.T) {
	data := &AppEnvVarData{App: &entity.App{Name: "web"}, EnvVars: []*EnvVar{usesHeldSecret()}}

	assert.Len(t, data.Errors(), 1)
}
