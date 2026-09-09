<b>[{{.ProjectName}}][{{.AppName}}] Deployment {{if .Succeeded}}✅ Succeeded{{else}}❌ Failed{{end}}</b>

<b>• Project:</b> {{.ProjectName}}
<b>• App:</b> {{.AppName}}
{{if .Method | eq "repo"}}
<b>• Repository:</b> <code>{{.RepoURL}}</code>
<b>• Branch/Ref:</b> <code>{{.RepoRef}}</code>
<b>• Commit Message:</b> <code>{{.CommitMsg}}</code>
<b>• Commit Author:</b> {{.CommitAuthor}}
{{else if .Method | eq "image"}}
<b>• Image:</b> <code>{{.Image}}</code>
{{end}}
<b>• Started At:</b> <code>{{.StartedAt}}</code>
<b>• Duration:</b> <code>{{.Duration}}</code>

🔗 <a href="{{.DashboardLink}}">Go to Dashboard</a>
