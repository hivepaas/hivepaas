{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json "System update " (outcome .Succeeded)}},
      "fields": [
        {
          "title": "Current Version",
          "value": {{json .CurrentVersion}},
          "short": true
        },
        {
          "title": "Target Version",
          "value": {{json .TargetVersion}},
          "short": true
        },
        {
          "title": "Started At",
          "value": {{json .StartedAt}},
          "short": true
        },
        {
          "title": "Duration",
          "value": {{json .Duration}},
          "short": true
        },
        {
          "title": "See task details",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}