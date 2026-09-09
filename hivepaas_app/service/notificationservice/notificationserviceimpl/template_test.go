package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	htmltemplate "html/template"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func TestDiscordAppDeploymentTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/app_deployment_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	tests := []struct {
		name string
		data notificationservice.TemplateDataAppDeployment
	}{
		{
			name: "deployment success - repo method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "My Project",
				AppName:       "My App",
				Succeeded:     true,
				Method:        "repo",
				RepoURL:       "https://github.com/user/repo",
				RepoRef:       "main",
				CommitMsg:     "initial commit with \"quotes\" and 'single' quotes",
				StartedAt:     time.Now(),
				Duration:      45 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard",
			},
		},
		{
			name: "deployment failure - image method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "Project A",
				AppName:       "App B",
				Succeeded:     false,
				Method:        "image",
				Image:         "nginx:latest",
				StartedAt:     time.Now(),
				Duration:      12 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard/project-a/app-b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tpl.Execute(&buf, tt.data)
			assert.NoError(t, err)

			output := buf.String()
			t.Logf("Generated JSON output:\n%s", output)

			// Try to parse output back to verify it is valid JSON
			var parsed map[string]any
			err = json.Unmarshal([]byte(output), &parsed)
			assert.NoError(t, err, "generated output should be valid JSON")

			// Verify fields in embeds
			embeds, ok := parsed["embeds"].([]any)
			assert.True(t, ok, "embeds should be an array")
			assert.Len(t, embeds, 1)

			embed, ok := embeds[0].(map[string]any)
			assert.True(t, ok, "embed should be an object")

			// Check color
			color := embed["color"].(float64)
			if tt.data.Succeeded {
				assert.Equal(t, float64(3066993), color)
			} else {
				assert.Equal(t, float64(15153724), color)
			}
		})
	}
}

func TestDiscordHealthcheckTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/healthcheck_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataHealthcheck{
		ProjectName:     "Test Project",
		AppName:         "Test App",
		Succeeded:       false,
		HealthcheckName: "ping-check",
		HealthcheckType: "http",
		StartedAt:       time.Now(),
		Duration:        1500 * time.Millisecond,
		Retries:         3,
		Expect:          "200 OK",
		Actual:          "500 Internal Server Error",
		DashboardLink:   "https://hivepaas.io/health",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestDiscordSchedTaskTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/sched_task_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSchedTask{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SchedJobName:  "db-backup",
		Schedule:      "0 0 * * *",
		StartedAt:     time.Now(),
		Duration:      15 * time.Second,
		Retries:       0,
		DashboardLink: "https://hivepaas.io/tasks",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestDiscordSSLExpiringTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/ssl_expiring_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLExpiring{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 0, 7),
		ExpireIn:      timeutil.Duration(7 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestDiscordSSLRenewalTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/ssl_renewal_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLRenewal{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 3, 0),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestDiscordSystemUpdateTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "discord/system_update_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSystemUpdate{
		Succeeded:      true,
		CurrentVersion: "v1.0.0",
		TargetVersion:  "v1.1.0",
		StartedAt:      time.Now(),
		Duration:       2 * time.Minute,
		DashboardLink:  "https://hivepaas.io/system",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestSlackAppDeploymentTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/app_deployment_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	tests := []struct {
		name string
		data notificationservice.TemplateDataAppDeployment
	}{
		{
			name: "deployment success - repo method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "My Project",
				AppName:       "My App",
				Succeeded:     true,
				Method:        "repo",
				RepoURL:       "https://github.com/user/repo",
				RepoRef:       "main",
				CommitMsg:     "initial commit with \"quotes\" and 'single' quotes",
				StartedAt:     time.Now(),
				Duration:      45 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard",
			},
		},
		{
			name: "deployment failure - image method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "Project A",
				AppName:       "App B",
				Succeeded:     false,
				Method:        "image",
				Image:         "nginx:latest",
				StartedAt:     time.Now(),
				Duration:      12 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard/project-a/app-b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tpl.Execute(&buf, tt.data)
			assert.NoError(t, err)

			output := buf.String()
			t.Logf("Generated JSON output:\n%s", output)

			var parsed map[string]any
			err = json.Unmarshal([]byte(output), &parsed)
			assert.NoError(t, err, "generated output should be valid JSON")

			attachments, ok := parsed["attachments"].([]any)
			assert.True(t, ok, "attachments should be an array")
			assert.Len(t, attachments, 1)

			attachment, ok := attachments[0].(map[string]any)
			assert.True(t, ok, "attachment should be an object")

			color := attachment["color"].(string)
			if tt.data.Succeeded {
				assert.Equal(t, "#2eb886", color)
			} else {
				assert.Equal(t, "#a30200", color)
			}
		})
	}
}

func TestSlackHealthcheckTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/healthcheck_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataHealthcheck{
		ProjectName:     "Test Project",
		AppName:         "Test App",
		Succeeded:       false,
		HealthcheckName: "ping-check",
		HealthcheckType: "http",
		StartedAt:       time.Now(),
		Duration:        1500 * time.Millisecond,
		Retries:         3,
		Expect:          "200 OK",
		Actual:          "500 Internal Server Error",
		DashboardLink:   "https://hivepaas.io/health",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestSlackSchedTaskTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/sched_task_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSchedTask{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SchedJobName:  "db-backup",
		Schedule:      "0 0 * * *",
		StartedAt:     time.Now(),
		Duration:      15 * time.Second,
		Retries:       0,
		DashboardLink: "https://hivepaas.io/tasks",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestSlackSSLExpiringTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/ssl_expiring_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLExpiring{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 0, 7),
		ExpireIn:      timeutil.Duration(7 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestSlackSSLRenewalTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/ssl_renewal_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLRenewal{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 3, 0),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestSlackSystemUpdateTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "slack/system_update_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSystemUpdate{
		Succeeded:      true,
		CurrentVersion: "v1.0.0",
		TargetVersion:  "v1.1.0",
		StartedAt:      time.Now(),
		Duration:       2 * time.Minute,
		DashboardLink:  "https://hivepaas.io/system",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err, "output should be valid JSON")
}

func TestTelegramAppDeploymentTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/app_deployment_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	tests := []struct {
		name string
		data notificationservice.TemplateDataAppDeployment
	}{
		{
			name: "deployment success - repo method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "My Project",
				AppName:       "My App",
				Succeeded:     true,
				Method:        "repo",
				RepoURL:       "https://github.com/user/repo",
				RepoRef:       "main",
				CommitMsg:     "initial commit with <brackets> and & ampersand",
				StartedAt:     time.Now(),
				Duration:      45 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard",
			},
		},
		{
			name: "deployment failure - image method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "Project A",
				AppName:       "App B",
				Succeeded:     false,
				Method:        "image",
				Image:         "nginx:latest",
				StartedAt:     time.Now(),
				Duration:      12 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard/project-a/app-b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tpl.Execute(&buf, tt.data)
			assert.NoError(t, err)

			output := buf.String()
			t.Logf("Generated Telegram HTML:\n%s", output)

			// Verify HTML safe escaping
			if tt.data.Method == "repo" {
				assert.Contains(t, output, "initial commit with &lt;brackets&gt; and &amp; ampersand")
			}
		})
	}
}

func TestTelegramHealthcheckTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/healthcheck_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataHealthcheck{
		ProjectName:     "Test Project",
		AppName:         "Test App",
		Succeeded:       false,
		HealthcheckName: "ping-check",
		HealthcheckType: "http",
		StartedAt:       time.Now(),
		Duration:        1500 * time.Millisecond,
		Retries:         3,
		Expect:          "200 OK",
		Actual:          "500 Internal Server Error",
		DashboardLink:   "https://hivepaas.io/health",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)
	t.Logf("Generated Telegram HTML:\n%s", buf.String())
}

func TestTelegramSchedTaskTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/sched_task_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSchedTask{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SchedJobName:  "db-backup",
		Schedule:      "0 0 * * *",
		StartedAt:     time.Now(),
		Duration:      15 * time.Second,
		Retries:       0,
		DashboardLink: "https://hivepaas.io/tasks",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)
	t.Logf("Generated Telegram HTML:\n%s", buf.String())
}

func TestTelegramSSLExpiringTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/ssl_expiring_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLExpiring{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 0, 7),
		ExpireIn:      timeutil.Duration(7 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)
	t.Logf("Generated Telegram HTML:\n%s", buf.String())
}

func TestTelegramSSLRenewalTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/ssl_renewal_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLRenewal{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().AddDate(0, 3, 0),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)
	t.Logf("Generated Telegram HTML:\n%s", buf.String())
}

func TestTelegramSystemUpdateTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "telegram/system_update_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSystemUpdate{
		Succeeded:      true,
		CurrentVersion: "v1.0.0",
		TargetVersion:  "v1.1.0",
		StartedAt:      time.Now(),
		Duration:       2 * time.Minute,
		DashboardLink:  "https://hivepaas.io/system",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)
	t.Logf("Generated Telegram HTML:\n%s", buf.String())
}

func TestLarkAppDeploymentTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/app_deployment_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	tests := []struct {
		name string
		data notificationservice.TemplateDataAppDeployment
	}{
		{
			name: "deployment success - repo method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "My Project",
				AppName:       "My App",
				Succeeded:     true,
				Method:        "repo",
				RepoURL:       "https://github.com/user/repo",
				RepoRef:       "main",
				CommitMsg:     "initial commit",
				StartedAt:     time.Now(),
				Duration:      45 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard",
			},
		},
		{
			name: "deployment failure - image method",
			data: notificationservice.TemplateDataAppDeployment{
				ProjectName:   "Project A",
				AppName:       "App B",
				Succeeded:     false,
				Method:        "image",
				Image:         "nginx:latest",
				StartedAt:     time.Now(),
				Duration:      12 * time.Second,
				DashboardLink: "https://hivepaas.io/dashboard/project-a/app-b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tpl.Execute(&buf, tt.data)
			assert.NoError(t, err)

			output := buf.String()
			t.Logf("Generated Lark JSON output:\n%s", output)

			var parsed map[string]any
			err = json.Unmarshal([]byte(output), &parsed)
			assert.NoError(t, err, "generated output should be valid JSON")
			assert.Equal(t, "interactive", parsed["msg_type"])
		})
	}
}

func TestLarkHealthcheckTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/healthcheck_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataHealthcheck{
		ProjectName:     "Test Project",
		AppName:         "Test App",
		Succeeded:       false,
		HealthcheckName: "ping-check",
		HealthcheckType: "http",
		Retries:         3,
		Expect:          "200 OK",
		Actual:          "500 Internal Server Error",
		StartedAt:       time.Now(),
		Duration:        1500 * time.Millisecond,
		DashboardLink:   "https://hivepaas.io/health",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "interactive", parsed["msg_type"])
}

func TestLarkSchedTaskTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/sched_task_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSchedTask{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SchedJobName:  "db-backup",
		Schedule:      "0 0 * * *",
		StartedAt:     time.Now(),
		Duration:      15 * time.Second,
		Retries:       0,
		DashboardLink: "https://hivepaas.io/tasks",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "interactive", parsed["msg_type"])
}

func TestLarkSSLExpiringTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/ssl_expiring_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLExpiring{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().Add(7 * 24 * time.Hour),
		ExpireIn:      timeutil.Duration(7 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "interactive", parsed["msg_type"])
}

func TestLarkSSLRenewalTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/ssl_renewal_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLRenewal{
		ProjectName:   "Test Project",
		AppName:       "Test App",
		Succeeded:     true,
		SSLName:       "my-cert",
		SSLType:       "Let's Encrypt",
		Domain:        "example.com",
		CreatedAt:     time.Now(),
		ExpireAt:      time.Now().Add(90 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "interactive", parsed["msg_type"])
}

func TestLarkSystemUpdateTemplate(t *testing.T) {
	tpl, err := texttemplate.ParseFS(assets.GetTemplatesFS(), "lark/system_update_notification.tpl")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSystemUpdate{
		Succeeded:      true,
		CurrentVersion: "v1.0.0",
		TargetVersion:  "v1.1.0",
		StartedAt:      time.Now(),
		Duration:       2 * time.Minute,
		DashboardLink:  "https://hivepaas.io/system",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "interactive", parsed["msg_type"])
}

func TestEmailSchedTaskTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/sched_task_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	t.Run("succeeded task", func(t *testing.T) {
		data := notificationservice.TemplateDataSchedTask{
			ProjectName:   "Alpha Project",
			AppName:       "Worker App",
			Succeeded:     true,
			SchedJobName:  "Database Backup",
			Schedule:      "every 1d",
			StartedAt:     time.Date(2026, 9, 9, 0, 31, 0, 0, time.UTC),
			Duration:      1500 * time.Millisecond,
			Retries:       0,
			DashboardLink: "https://hivepaas.io/tasks/123",
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, data)
		assert.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "Scheduled task succeeded")
		assert.Contains(t, out, "badge-succeeded")
		assert.Contains(t, out, "Database Backup")
		assert.Contains(t, out, "Sep 09, 2026, 00:31:00 UTC")
		assert.Contains(t, out, "https://hivepaas.io/tasks/123")
		assert.NotContains(t, out, "Error Details")
	})

	t.Run("failed task with last error", func(t *testing.T) {
		data := notificationservice.TemplateDataSchedTask{
			ProjectName:   "Alpha Project",
			AppName:       "Worker App",
			Succeeded:     false,
			SchedJobName:  "System backup job",
			Schedule:      "every 1d",
			StartedAt:     time.Date(2026, 9, 9, 0, 31, 0, 0, time.UTC),
			Duration:      6 * time.Millisecond,
			Retries:       1,
			LastError:     "mysqldump: Got error 2002: Can't connect to server",
			DashboardLink: "https://hivepaas.io/tasks/456",
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, data)
		assert.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "Scheduled task failed")
		assert.Contains(t, out, "badge-failed")
		assert.Contains(t, out, "Error Details")
		assert.Contains(t, out, "mysqldump: Got error 2002: Can&#39;t connect to server")
		assert.Contains(t, out, "Sep 09, 2026, 00:31:00 UTC")
		assert.Contains(t, out, "https://hivepaas.io/tasks/456")
	})
}

func TestEmailAppDeploymentTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/app_deployment_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataAppDeployment{
		ProjectName:   "Alpha Project",
		AppName:       "Web App",
		Succeeded:     true,
		Method:        "repo",
		RepoURL:       "https://github.com/hivepaas/hivepaas",
		RepoRef:       "main",
		CommitMsg:     "feat: improve email templates",
		CommitAuthor:  "Dev",
		StartedAt:     time.Date(2026, 9, 9, 1, 0, 0, 0, time.UTC),
		Duration:      45 * time.Second,
		DashboardLink: "https://hivepaas.io/deployments/123",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Deployment succeeded")
	assert.Contains(t, out, "badge-succeeded")
	assert.Contains(t, out, "Sep 09, 2026, 01:00:00 UTC")
	assert.Contains(t, out, "https://hivepaas.io/deployments/123")
}

func TestEmailHealthcheckTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/healthcheck_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataHealthcheck{
		ProjectName:     "Alpha Project",
		AppName:         "API Service",
		Succeeded:       false,
		HealthcheckName: "ping-check",
		HealthcheckType: "http",
		Expect:          "200 OK",
		Actual:          "503 Service Unavailable",
		StartedAt:       time.Date(2026, 9, 9, 1, 0, 0, 0, time.UTC),
		Duration:        200 * time.Millisecond,
		Retries:         3,
		DashboardLink:   "https://hivepaas.io/health/123",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Healthcheck failed")
	assert.Contains(t, out, "badge-failed")
	assert.Contains(t, out, "Check Result Mismatch")
	assert.Contains(t, out, "503 Service Unavailable")
	assert.Contains(t, out, "Sep 09, 2026, 01:00:00 UTC")
}

func TestEmailSSLExpiringTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/ssl_expiring_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSSLExpiring{
		ProjectName:   "Alpha Project",
		AppName:       "Web App",
		SSLName:       "Production Cert",
		SSLType:       "Let's Encrypt",
		Domain:        "app.example.com",
		CreatedAt:     time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		ExpireAt:      time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC),
		ExpireIn:      timeutil.Duration(7 * 24 * time.Hour),
		DashboardLink: "https://hivepaas.io/ssl/123",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "SSL expiring in")
	assert.Contains(t, out, "badge-warning")
	assert.Contains(t, out, "Jun 09, 2026, 00:00:00 UTC")
	assert.Contains(t, out, "Sep 16, 2026, 00:00:00 UTC")
	assert.Contains(t, out, "https://hivepaas.io/ssl/123")
}

func TestEmailSSLRenewalTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/ssl_renewal_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	t.Run("renewal succeeded", func(t *testing.T) {
		data := notificationservice.TemplateDataSSLRenewal{
			ProjectName:   "Alpha Project",
			AppName:       "Web App",
			Succeeded:     true,
			SSLName:       "Production Cert",
			SSLType:       "Let's Encrypt",
			Domain:        "app.example.com",
			CreatedAt:     time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
			ExpireAt:      time.Date(2026, 12, 9, 0, 0, 0, 0, time.UTC),
			NextRenewalIn: timeutil.Duration(60 * 24 * time.Hour),
			DashboardLink: "https://hivepaas.io/ssl/123",
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, data)
		assert.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "SSL renewal succeeded")
		assert.Contains(t, out, "badge-succeeded")
		assert.Contains(t, out, "Jun 09, 2026, 00:00:00 UTC")
	})

	t.Run("renewal failed", func(t *testing.T) {
		data := notificationservice.TemplateDataSSLRenewal{
			ProjectName:   "Alpha Project",
			AppName:       "Web App",
			Succeeded:     false,
			SSLName:       "Production Cert",
			SSLType:       "Let's Encrypt",
			Domain:        "app.example.com",
			CreatedAt:     time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
			ExpireAt:      time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
			NextRenewalIn: timeutil.Duration(1 * 24 * time.Hour),
			DashboardLink: "https://hivepaas.io/ssl/123",
		}

		var buf bytes.Buffer
		err = tpl.Execute(&buf, data)
		assert.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "SSL renewal failed")
		assert.Contains(t, out, "badge-failed")
	})
}

func TestEmailSystemUpdateTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/system_update_notification.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := notificationservice.TemplateDataSystemUpdate{
		Succeeded:      true,
		CurrentVersion: "v1.2.0",
		TargetVersion:  "v1.3.0",
		StartedAt:      time.Date(2026, 9, 9, 2, 0, 0, 0, time.UTC),
		Duration:       45 * time.Second,
		DashboardLink:  "https://hivepaas.io/system",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "System update succeeded")
	assert.Contains(t, out, "badge-succeeded")
	assert.Contains(t, out, "v1.2.0")
	assert.Contains(t, out, "v1.3.0")
	assert.Contains(t, out, "Sep 09, 2026, 02:00:00 UTC")
}

func TestEmailPasswordResetTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/password_reset.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := struct {
		ResetPasswordLink string
	}{
		ResetPasswordLink: "https://hivepaas.io/reset?token=xyz",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Reset your password")
	assert.Contains(t, out, "badge-security")
	assert.Contains(t, out, "https://hivepaas.io/reset?token=xyz")
}

func TestEmailUserInviteTemplate(t *testing.T) {
	tpl, err := htmltemplate.ParseFS(assets.GetTemplatesFS(), "email/user_invite.html")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := struct {
		InviterName    string
		UserSignupLink string
	}{
		InviterName:    "Alice Smith",
		UserSignupLink: "https://hivepaas.io/invite?code=abc",
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Join HivePaaS")
	assert.Contains(t, out, "badge-invite")
	assert.Contains(t, out, "Alice Smith")
	assert.Contains(t, out, "https://hivepaas.io/invite?code=abc")
}
