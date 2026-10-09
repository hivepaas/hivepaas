{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json (scope .ProjectName .AppName) " Healthcheck " (outcome .Succeeded)}},
      "fields": [
        {{if .ProjectName | ne ""}}{
          "title": "Project",
          "value": {{json .ProjectName}},
          "short": true
        },{{end}}
        {{if .AppName | ne ""}}{
          "title": "App",
          "value": {{json .AppName}},
          "short": true
        },{{end}}
        {
          "title": "Name",
          "value": {{json .HealthcheckName}},
          "short": true
        },
        {
          "title": "Type",
          "value": {{json .HealthcheckType}},
          "short": true
        },
        {
          "title": "Retries",
          "value": {{json .Retries}},
          "short": true
        },
        {{if not .Succeeded}}{
          "title": "Expect",
          "value": {{json .Expect}},
          "short": false
        },
        {
          "title": "Actual",
          "value": {{json .Actual}},
          "short": false
        },{{end}}{
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
