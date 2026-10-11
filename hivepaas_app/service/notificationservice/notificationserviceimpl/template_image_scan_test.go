package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	htmltemplate "html/template"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func imageScanTemplateData() notificationservice.TemplateDataImageScan {
	return notificationservice.TemplateDataImageScan{
		BaseTemplateData: notificationservice.BaseTemplateData{
			Title: "[System] 2 apps run new fixable vulnerabilities",
		},
		Apps: []*notificationservice.ImageScanApp{
			{ProjectName: "Shop", EnvName: "production", AppName: "api \"v2\"", Critical: 2, High: 1,
				Link: "https://hivepaas.io/projects/P1/production/apps/A1/security"},
			{ProjectName: "Shop", EnvName: "staging", AppName: "web", High: 3,
				Link: "https://hivepaas.io/projects/P1/staging/apps/A2/security"},
		},
		DashboardLink: "https://hivepaas.io/system/image-scanning",
	}
}

// The daily scan's report is valid JSON on every JSON channel, and names each
// app with what is new in it.
func TestImageScanJSONTemplates(t *testing.T) {
	for _, path := range []string{"slack/image_scan_notification.tpl", "discord/image_scan_notification.tpl",
		"lark/image_scan_notification.tpl"} {
		t.Run(path, func(t *testing.T) {
			tpl, err := parseJSONTemplate(path)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			var buf bytes.Buffer
			assert.NoError(t, tpl.Execute(&buf, imageScanTemplateData()))
			var parsed map[string]any
			assert.NoError(t, json.Unmarshal(buf.Bytes(), &parsed), buf.String())
			assert.Contains(t, buf.String(), `Shop / production / api \"v2\": 2 CRITICAL, 1 HIGH`)
			assert.Contains(t, buf.String(), `Shop / staging / web: 3 HIGH`)
			assert.Contains(t, buf.String(), "https://hivepaas.io/system/image-scanning")
		})
	}
}

func TestImageScanHTMLTemplates(t *testing.T) {
	for _, path := range []string{"telegram/image_scan_notification.tpl", "email/image_scan_notification.html"} {
		t.Run(path, func(t *testing.T) {
			tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), path)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			var buf bytes.Buffer
			assert.NoError(t, tpl.Execute(&buf, imageScanTemplateData()))
			out := buf.String()
			assert.Contains(t, out, "2 apps run new fixable vulnerabilities")
			assert.Contains(t, out, "Shop / production / api &#34;v2&#34;")
			assert.Contains(t, out, "2 CRITICAL, 1 HIGH")
			assert.Contains(t, out, "https://hivepaas.io/system/image-scanning")
		})
	}
}
