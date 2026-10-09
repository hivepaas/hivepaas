{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json (scope .ProjectName .AppName) " Scheduled task " (outcome .Succeeded)}},
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
          "title": "Scheduled Job",
          "value": {{json .SchedJobName}},
          "short": true
        },
        {
          "title": "Schedule",
          "value": {{json .Schedule}},
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
          "title": "Retries",
          "value": {{json .Retries}},
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