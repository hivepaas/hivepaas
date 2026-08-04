{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Scheduled task {{if .Succeeded}}succeeded{{else}}failed{{end}}",
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
          "title": "Scheduled Job",
          "value": {{printf "%q" .SchedJobName}},
          "short": true
        },
        {
          "title": "Schedule",
          "value": {{printf "%q" .Schedule}},
          "short": true
        },
        {
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
          "title": "Retries",
          "value": "{{.Retries}}",
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