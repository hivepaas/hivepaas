{
  "attachments": [
    {
      "color": "#f1c40f",
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL expiring in {{.ExpireIn}}",
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
          "value": {{printf "%q" .SSLName}},
          "short": true
        },
        {
          "title": "Type",
          "value": {{printf "%q" .SSLType}},
          "short": true
        },
        {
          "title": "Domain",
          "value": {{printf "%q" .Domain}},
          "short": true
        },
        {
          "title": "Created At",
          "value": {{printf "%q" .CreatedAt}},
          "short": true
        },
        {
          "title": "Expire At",
          "value": {{printf "%q" .ExpireAt}},
          "short": true
        },
        {
          "title": "Expire In",
          "value": {{printf "%q" .ExpireIn}},
          "short": true
        },
        {
          "title": "See object details",
          "value": "<{{.DashboardLink}}|Go to Dashboard>",
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}