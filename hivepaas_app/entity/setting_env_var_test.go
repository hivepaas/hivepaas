package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// What reaches a container is text, so a value written as a scalar is read as
// the text it would print as - which is how a template parameter that asks a
// yes-or-no question, or a number, arrives here.
func TestEnvVarValueWrittenAsAScalar(t *testing.T) {
	cases := map[string]struct {
		body string
		want string
	}{
		"text":            {`{"k": "MODE", "v": "debug"}`, "debug"},
		"true":            {`{"k": "MODE", "v": true}`, "true"},
		"false":           {`{"k": "MODE", "v": false}`, "false"},
		"a whole number":  {`{"k": "MODE", "v": 8080}`, "8080"},
		"a decimal":       {`{"k": "MODE", "v": 1.5}`, "1.5"},
		"null":            {`{"k": "MODE", "v": null}`, ""},
		"nothing at all":  {`{"k": "MODE"}`, ""},
		"text that lies":  {`{"k": "MODE", "v": "true"}`, "true"},
		"an empty string": {`{"k": "MODE", "v": ""}`, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			envVar := &EnvVar{}
			assert.NoError(t, json.Unmarshal([]byte(tc.body), envVar))
			assert.Equal(t, "MODE", envVar.Key)
			assert.Equal(t, tc.want, envVar.Value)
		})
	}
}

func TestEnvVarKeepsReadingTheRestOfTheFields(t *testing.T) {
	envVar := &EnvVar{}

	assert.NoError(t, json.Unmarshal([]byte(`{"k": "MODE", "v": 1, "build": true, "shared": true}`), envVar))

	assert.True(t, envVar.IsBuild)
	assert.True(t, envVar.IsShared)
	assert.Equal(t, "1", envVar.Value)
}

// A field nobody defined is still a mistake worth reporting: a template that
// writes "value" instead of "v" would otherwise set nothing and say nothing.
func TestEnvVarRefusesWhatItCannotBe(t *testing.T) {
	for _, body := range []string{
		`{"k": "MODE", "v": {"a": 1}}`,
		`{"k": "MODE", "v": ["a"]}`,
		`{"k": "MODE", "value": "debug"}`,
	} {
		assert.Error(t, json.Unmarshal([]byte(body), &EnvVar{}), body)
	}
}

// Whatever it was written as, it is stored as text.
func TestEnvVarIsWrittenBackAsText(t *testing.T) {
	envVar := &EnvVar{}
	assert.NoError(t, json.Unmarshal([]byte(`{"k": "DEBUG", "v": true}`), envVar))

	encoded, err := json.Marshal(envVar)

	assert.NoError(t, err)
	assert.JSONEq(t, `{"k": "DEBUG", "v": "true"}`, string(encoded))
}
