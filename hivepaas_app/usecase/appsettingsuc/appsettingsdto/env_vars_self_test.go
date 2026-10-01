package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envself"
)

func TestEnvSelfSuggestionsEngineIsOneOfTheList(t *testing.T) {
	req := &GetEnvSelfSuggestionsReq{ProjectID: "01HZZZZZZZZZZZZZZZZZZZZZZZ",
		ProjectEnvID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", AppID: "01HZZZZZZZZZZZZZZZZZZZZZZZ"}
	assert.Empty(t, req.Validate(), "App Kind's when not given")
	req.Engine = "postgres"
	assert.Empty(t, req.Validate())
	req.Engine = "cassandra"
	assert.NotEmpty(t, req.Validate())
}

// With no engine to suggest for, the list is still answered, to choose from.
func TestEnvSelfSuggestionsWithoutAnEngine(t *testing.T) {
	resp := TransformEnvSelfSuggestion(nil, nil)
	assert.Len(t, resp.Engines, len(envself.Engines))
	assert.Nil(t, resp.Engine)
	assert.NotNil(t, resp.Vars)
	assert.NotNil(t, resp.Warnings)
}
