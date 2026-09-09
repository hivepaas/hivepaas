{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": "System update {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "fields": [
        {
          "title": "Current Version",
          "value": {{printf "%q" .CurrentVersion}},
          "short": true
        },
        {
          "title": "Target Version",
          "value": {{printf "%q" .TargetVersion}},
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
          "title": "See task details",
          "value": "<{{.DashboardLink}}|Go to Dashboard>",
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}