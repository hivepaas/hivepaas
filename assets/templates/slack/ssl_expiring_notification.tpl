{
  "attachments": [
    {
      "color": "#f1c40f",
      "title": {{json (scope .ProjectName .AppName) " SSL expiring in " .ExpireIn}},
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
          "value": {{json .SSLName}},
          "short": true
        },
        {
          "title": "Type",
          "value": {{json .SSLType}},
          "short": true
        },
        {
          "title": "Domain",
          "value": {{json .Domain}},
          "short": true
        },
        {
          "title": "Created At",
          "value": {{json .CreatedAt}},
          "short": true
        },
        {
          "title": "Expire At",
          "value": {{json .ExpireAt}},
          "short": true
        },
        {
          "title": "Expire In",
          "value": {{json .ExpireIn}},
          "short": true
        },
        {
          "title": "See object details",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}