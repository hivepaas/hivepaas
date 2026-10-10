{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json "[" .ProjectName "][" .AppName "] Clone " (outcome .Succeeded)}},
      "fields": [
        {
          "title": "Project",
          "value": {{json .ProjectName}},
          "short": true
        },
        {
          "title": "App",
          "value": {{json .AppName}},
          "short": true
        },
        {
          "title": "Clone",
          "value": {{json .CloneName " in " .CloneEnv}},
          "short": false
        },
        {{if .Reason}}{
          "title": "Reason",
          "value": {{json .Reason}},
          "short": false
        },
        {{end}}{
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
          "title": "See clone details",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}
