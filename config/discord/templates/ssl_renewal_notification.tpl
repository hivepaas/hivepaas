**{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL renewal {{if .Succeeded}}succeeded{{else}}failed{{end}}**
{{if .ProjectName | ne ""}}> Project: `{{.ProjectName}}`{{end}}
{{if .AppName | ne ""}}> App: `{{.AppName}}`{{end}}
> Name: `{{.SSLName}}`
> Type: `{{.SSLType}}`
> Domain: `{{.Domain}}`
> Created at: `{{.CreatedAt}}`
> Expire at: `{{.ExpireAt}}`
{{if .NextRenewalIn | gt 0}}> Next renewal in: `{{.NextRenewalIn}}`{{end}}
> See task details: {{.DashboardLink}}