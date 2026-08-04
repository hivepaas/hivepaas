{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Healthcheck {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "fields": [
        {{if .ProjectName | ne ""}}{
          "title": "Project",
          "value": {{printf "%q" .ProjectName}},
          "short": true
        },{{end}}
        {{if .AppName | ne ""}}{
          "title": "App",
          "value": {{printf "%q" .AppName}},
          "short": true
        },{{end}}
        {
          "title": "Name",
          "value": {{printf "%q" .HealthcheckName}},
          "short": true
        },
        {
          "title": "Type",
          "value": {{printf "%q" .HealthcheckType}},
          "short": true
        },
        {
          "title": "Retries",
          "value": "{{.Retries}}",
          "short": true
        },
        {{if not .Succeeded}}{
          "title": "Expect",
          "value": {{printf "%q" .Expect}},
          "short": false
        },
        {
          "title": "Actual",
          "value": {{printf "%q" .Actual}},
          "short": false
        },{{end}}{
          "title": "Started At",
          "value": {{printf "%q" .StartedAt}},
          "short": true
        },
        {
          "title": "Duration",
          "value": {{printf "%q" .Duration}},
          "short": true
        },
        {
          "title": "See task details",
          "value": "<{{.DashboardLink}}|Go to Dashboard>",
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}
