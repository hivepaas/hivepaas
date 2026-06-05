{
  "embeds": [
    {
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL expiring in {{.ExpireIn}}",
      "color": 15844367,
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
          "value": {{printf "%q" .SSLName}},
          "inline": true
        },
        {
          "name": "Type",
          "value": {{printf "%q" .SSLType}},
          "inline": true
        },
        {
          "name": "Domain",
          "value": {{printf "%q" .Domain}},
          "inline": true
        },
        {
          "name": "Created At",
          "value": {{printf "%q" .CreatedAt}},
          "inline": true
        },
        {
          "name": "Expire At",
          "value": {{printf "%q" .ExpireAt}},
          "inline": true
        },
        {
          "name": "Expire In",
          "value": {{printf "%q" .ExpireIn}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See object details",
          "value": "[Go to Dashboard]({{.DashboardLink}})",
          "inline": false
        }
      ]
    }
  ]
}