package registryserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSearchResponseSumsWhatIsStored(t *testing.T) {
	body := []byte(`{"data":{"RepoListWithNewestImage":{"Results":[
		{"Name":"hivepaas/api","Size":"24148826"},
		{"Name":"hivepaas/web","Size":"4139411"}]}}}`)

	repos, stored, err := parseSearchResponse(body)

	assert.NoError(t, err)
	assert.Equal(t, 2, repos)
	assert.Equal(t, int64(28288237), stored)
}

// An empty registry is a normal state, not a failure to read one.
func TestParseSearchResponseOnAnEmptyRegistry(t *testing.T) {
	repos, stored, err := parseSearchResponse([]byte(`{"data":{"RepoListWithNewestImage":{"Results":[]}}}`))

	assert.NoError(t, err)
	assert.Zero(t, repos)
	assert.Zero(t, stored)
}

// The search extension answers a bad query with 200, so the body is the only
// place a failure shows up.
func TestParseSearchResponseReportsGraphQLErrors(t *testing.T) {
	_, _, err := parseSearchResponse([]byte(`{"errors":[{"message":"Cannot query field \"Name\""}]}`))

	assert.Error(t, err)
}
