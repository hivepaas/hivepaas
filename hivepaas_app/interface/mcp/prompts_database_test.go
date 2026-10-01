package mcp

import (
	"context"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func promptText(t *testing.T, session *mcpsdk.ClientSession, name string, args map[string]string) string {
	t.Helper()
	res, err := session.GetPrompt(context.Background(), &mcpsdk.GetPromptParams{Name: name, Arguments: args})
	if !assert.NoError(t, err, name) || !assert.Len(t, res.Messages, 1, name) {
		t.FailNow()
	}
	return res.Messages[0].Content.(*mcpsdk.TextContent).Text
}

func TestTheDatabasesGuideIsAResource(t *testing.T) {
	_, _, session := clusterSession(t)
	res, err := session.ReadResource(context.Background(), &mcpsdk.ReadResourceParams{URI: databasesURI})
	if assert.NoError(t, err) && assert.Len(t, res.Contents, 1) {
		text := res.Contents[0].Text
		assert.Contains(t, text, "get_env_link_suggestions")
		assert.Contains(t, text, "get_env_self_suggestions")
		assert.Contains(t, text, "MySQL and MariaDB do\n  not")
	}
}

// Each database prompt names the tools it takes and the guide, and ends as the
// key and the server allow: a plan to agree to, or what to change and where.
func TestDatabasePrompts(t *testing.T) {
	app := map[string]string{"project": "shop", "env": "prod", "app": "api"}
	withDB := map[string]string{"project": "shop", "env": "prod", "app": "api", "database": "db"}
	for name, want := range map[string][]string{
		"connect_app_to_database": {"get_env_link_suggestions", "kind env-vars"},
		"run_database_from_image": {"get_env_self_suggestions", "initOnly", "never one you make up"},
		"expose_database":         {"kind routing", "protocol tcp", "MySQL or MariaDB", "firewall"},
	} {
		args := app
		if name == "connect_app_to_database" {
			args = withDB
		}
		w := newWriteWorld(t)
		text := promptText(t, w.session(t, "key1"), name, args)
		assert.Contains(t, text, "api of shop, env prod", name)
		assert.Contains(t, text, databasesURI, name)
		for _, s := range want {
			assert.Contains(t, text, s, name)
		}
		assert.Contains(t, text, "apply it only once I agree", name)

		text = promptText(t, w.session(t, "reader"), name, args)
		assert.Contains(t, text, "Change nothing: this key may not", name)
		assert.NotContains(t, text, "apply it only once I agree", name)
	}

	_, _, session := clusterSession(t) // changes off on this server
	text := promptText(t, session, "expose_database", app)
	assert.Contains(t, text, "changes are off on this server")
	text = promptText(t, session, "connect_app_to_database", app)
	assert.Contains(t, text, "list_env_link_targets, and ask me which", "no database named")
}
