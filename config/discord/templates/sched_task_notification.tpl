**{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Scheduled task {{if .Succeeded}}succeeded{{else}}failed{{end}}**
{{if .ProjectName | ne ""}}> Project: `{{.ProjectName}}`{{end}}
{{if .AppName | ne ""}}> App: `{{.AppName}}`{{end}}
> Scheduled job: `{{.SchedJobName}}`
> Schedule: `{{.Schedule}}`
> Created at: `{{.CreatedAt}}`
> Started at: `{{.StartedAt}}`
> Duration: `{{.Duration}}`
> Retries: `{{.Retries}}`
> See task details: {{.DashboardLink}}
