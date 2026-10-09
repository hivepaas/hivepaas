package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	htmltemplate "html/template"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

type executor interface {
	Execute(w io.Writer, data any) error
}

// A failed deployment tells why on every channel, its words kept as they are:
// quotes and new lines too, which a JSON body must escape. A succeeded one has
// no reason to tell.
func TestAppDeploymentTemplatesTellTheReason(t *testing.T) {
	reason := `pulling traefik/whoami:e2e: manifest "e2e" not found` + "\nERR_NOT_FOUND"
	failed := notificationservice.TemplateDataAppDeployment{
		ProjectName:   `Shop "eu"`,
		AppName:       "web",
		Method:        "repo",
		RepoURL:       "https://github.com/acme/shop",
		RepoRef:       "refs/heads/main",
		CommitMsg:     "Say \"hello\"\n\nand a body",
		CommitAuthor:  "Dev",
		StartedAt:     time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC),
		Duration:      12 * time.Second,
		DashboardLink: "https://hivepaas.example/deployments/1",
		Reason:        reason,
	}
	succeeded := failed
	succeeded.Succeeded, succeeded.Reason = true, ""

	parse := func(name string) executor {
		var tpl executor
		var err error
		if strings.HasSuffix(name, ".html") || strings.HasPrefix(name, "telegram/") {
			tpl, err = htmltemplate.ParseFS(assets.GetTemplatesFS(), name)
		} else {
			tpl, err = parseJSONTemplate(name)
		}
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		return tpl
	}
	render := func(name string, data notificationservice.TemplateDataAppDeployment) string {
		var buf bytes.Buffer
		if err := parse(name).Execute(&buf, data); err != nil {
			t.Fatalf("rendering %s: %v", name, err)
		}
		return buf.String()
	}

	for _, name := range []string{
		"slack/app_deployment_notification.tpl",
		"discord/app_deployment_notification.tpl",
		"lark/app_deployment_notification.tpl",
	} {
		t.Run(name, func(t *testing.T) {
			for _, data := range []notificationservice.TemplateDataAppDeployment{failed, succeeded} {
				out := render(name, data)
				var parsed any
				if err := json.Unmarshal([]byte(out), &parsed); err != nil {
					t.Fatalf("not JSON: %v\n%s", err, out)
				}
				// What a reader sees is the reason, unescaped once parsed.
				texts, _ := json.Marshal(parsed)
				decoded := string(texts)
				if data.Succeeded {
					assert.NotContains(t, decoded, "Reason")
				} else {
					assert.Contains(t, decoded, "Reason")
					wantReason, _ := json.Marshal(reason)
					assert.Contains(t, decoded, strings.Trim(string(wantReason), `"`))
				}
			}
		})
	}

	t.Run("telegram", func(t *testing.T) {
		out := render("telegram/app_deployment_notification.tpl", failed)
		assert.Contains(t, out, "<b>• Reason:</b> <code>pulling traefik/whoami:e2e: manifest &#34;e2e&#34; not found")
		assert.NotContains(t, render("telegram/app_deployment_notification.tpl", succeeded), "Reason")
	})
	t.Run("email", func(t *testing.T) {
		out := render("email/app_deployment_notification.html", failed)
		assert.Contains(t, out, `<div class="err-title">Reason</div>`)
		assert.Contains(t, out, "manifest &#34;e2e&#34; not found")
		assert.NotContains(t, render("email/app_deployment_notification.html", succeeded), `class="err-box"`)
	})
}
