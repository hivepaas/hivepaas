<b>{{if .ProjectName}}[{{.ProjectName}}]{{if .AppName}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Healthcheck {{if .Succeeded}}✅ Succeeded{{else}}❌ Failed{{end}}</b>

{{if .ProjectName}}<b>• Project:</b> {{.ProjectName}}
{{end}}{{if .AppName}}<b>• App:</b> {{.AppName}}
{{end}}<b>• Name:</b> <code>{{.HealthcheckName}}</code>
<b>• Type:</b> <code>{{.HealthcheckType}}</code>
<b>• Retries:</b> <code>{{.Retries}}</code>
{{if not .Succeeded}}<b>• Expect:</b> <code>{{.Expect}}</code>
<b>• Actual:</b> <code>{{.Actual}}</code>
{{end}}<b>• Started At:</b> <code>{{.StartedAt}}</code>
<b>• Duration:</b> <code>{{.Duration}}</code>

🔗 <a href="{{.DashboardLink}}">Go to Dashboard</a>
