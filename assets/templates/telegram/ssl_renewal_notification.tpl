<b>{{if .ProjectName}}[{{.ProjectName}}]{{if .AppName}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL renewal {{if .Succeeded}}✅ Succeeded{{else}}❌ Failed{{end}}</b>

{{if .ProjectName}}<b>• Project:</b> {{.ProjectName}}
{{end}}{{if .AppName}}<b>• App:</b> {{.AppName}}
{{end}}<b>• Name:</b> <code>{{.SSLName}}</code>
<b>• Type:</b> <code>{{.SSLType}}</code>
<b>• Domain:</b> <code>{{.Domain}}</code>
<b>• Created At:</b> <code>{{.CreatedAt}}</code>
<b>• Expire At:</b> <code>{{.ExpireAt}}</code>
{{if .NextRenewalIn | gt 0}}<b>• Next Renewal In:</b> <code>{{.NextRenewalIn}}</code>
{{end}}
🔗 <a href="{{.DashboardLink}}">Go to Dashboard</a>
