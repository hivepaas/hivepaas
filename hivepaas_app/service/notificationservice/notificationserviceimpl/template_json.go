package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	texttemplate "text/template"

	"github.com/hivepaas/hivepaas/assets"
)

// systemTitlePrefix is a title's prefix for what is the installation's own.
const systemTitlePrefix = "[System]"

// jsonTemplateFuncs are what the templates of a JSON body - Slack's, Discord's,
// Lark's - write their values with. A value is anybody's: a project's or an
// app's name takes any character, a commit message has quotes and new lines, a
// health check's answer may hold control characters. Written into a JSON string
// as it is, any of them breaks the body, and the notification is not sent.
var jsonTemplateFuncs = texttemplate.FuncMap{
	// json is its arguments, put together as print puts them, as a JSON string.
	"json": jsonString,
	// scope is the title's prefix: [project][app], [project], or [System].
	"scope": func(project, app string) string {
		switch {
		case project == "":
			return systemTitlePrefix
		case app == "":
			return "[" + project + "]"
		default:
			return "[" + project + "][" + app + "]"
		}
	},
	// outcome is how a run ended, as a title says it.
	"outcome": func(succeeded bool) string {
		if succeeded {
			return "succeeded"
		}
		return "failed"
	},
}

func jsonString(parts ...any) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	// Kept as they are, < and > read better in a body logged or inspected:
	// nothing here is ever HTML.
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(fmt.Sprint(parts...)); err != nil {
		return "", err //nolint:wrapcheck
	}
	return string(bytes.TrimSuffix(buf.Bytes(), []byte("\n"))), nil
}

// parseJSONTemplate parses a template of a JSON body, with what it writes its
// values with.
func parseJSONTemplate(file string) (*texttemplate.Template, error) {
	return texttemplate.New(path.Base(file)).Funcs(jsonTemplateFuncs). //nolint:wrapcheck
										ParseFS(assets.GetTemplatesFS(), file)
}
