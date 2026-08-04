{
  "embeds": [
    {
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Healthcheck {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {{if .ProjectName | ne ""}}{
          "name": "Project",
          "value": {{printf "%q" .ProjectName}},
          "inline": true
        },{{end}}
        {{if .AppName | ne ""}}{
          "name": "App",
          "value": {{printf "%q" .AppName}},
          "inline": true
        },{{end}}
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Name",
          "value": {{printf "%q" .HealthcheckName}},
          "inline": true
        },
        {
          "name": "Type",
          "value": {{printf "%q" .HealthcheckType}},
          "inline": true
        },
        {
          "name": "Retries",
          "value": "{{.Retries}}",
          "inline": true
        },
        {{if not .Succeeded}}{
          "name": "Expect",
          "value": {{printf "%q" .Expect}},
          "inline": false
        },
        {
          "name": "Actual",
          "value": {{printf "%q" .Actual}},
          "inline": false
        },{{end}}{
          "name": "Started At",
          "value": {{printf "%q" .StartedAt}},
          "inline": true
        },
        {
          "name": "Duration",
          "value": {{printf "%q" .Duration}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See task details",
          "value": "[Go to Dashboard]({{.DashboardLink}})",
          "inline": false
        }
      ]
    }
  ]
}
