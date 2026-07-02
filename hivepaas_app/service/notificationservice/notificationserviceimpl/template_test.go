package notificationserviceimpl

import (
	"bytes"
	"encoding/json"
	htmltemplate "html/template"
	"path/filepath"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
)

func TestDiscordAppDeploymentTemplate(t *testing.T) {
	// Locate template file relative to this test file
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "app_deployment_notification.tpl")

	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "healthcheck_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "sched_task_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "ssl_expiring_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "ssl_renewal_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "discord", "templates", "system_update_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "app_deployment_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "healthcheck_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "sched_task_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "ssl_expiring_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "ssl_renewal_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "slack", "templates", "system_update_notification.tpl")
	tpl, err := texttemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "app_deployment_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "healthcheck_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "sched_task_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "ssl_expiring_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "ssl_renewal_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
	tplPath := filepath.Join("..", "..", "..", "..", "config", "telegram", "templates", "system_update_notification.tpl")
	tpl, err := htmltemplate.ParseFiles(tplPath)
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
