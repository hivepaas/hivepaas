<b>[{{.ProjectName}}][{{.AppName}}] Clone {{if .Succeeded}}✅ Succeeded{{else}}❌ Failed{{end}}</b>

<b>• Project:</b> {{.ProjectName}}
<b>• App:</b> {{.AppName}}
<b>• Clone:</b> {{.CloneName}} in {{.CloneEnv}}
{{if .Reason}}<b>• Reason:</b> <code>{{.Reason}}</code>
{{end}}
<b>• Started At:</b> <code>{{.StartedAt}}</code>
<b>• Duration:</b> <code>{{.Duration}}</code>

🔗 <a href="{{.DashboardLink}}">Go to Dashboard</a>
