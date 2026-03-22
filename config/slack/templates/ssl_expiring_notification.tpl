*{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL expiring in {{.ExpireIn}}*
{{if .ProjectName | ne ""}}> Project: `{{.ProjectName}}`{{end}}
{{if .AppName | ne ""}}> App: `{{.AppName}}`{{end}}
> Name: `{{.SSLName}}`
> Type: `{{.SSLType}}`
> Domain: `{{.Domain}}`
> Created at: `{{.CreatedAt}}`
> Expire at: `{{.ExpireAt}}`
{{if .ExpireIn | gt 0}}> Expire in: `{{.ExpireIn}}`{{end}}
> See object details: {{.DashboardLink}}