{
  "embeds": [
    {
      "title": {{json (scope .ProjectName .AppName) " SSL expiring in " .ExpireIn}},
      "color": 15844367,
      "fields": [
        {{if .ProjectName | ne ""}}{
          "name": "Project",
          "value": {{json .ProjectName}},
          "inline": true
        },{{end}}
        {{if .AppName | ne ""}}{
          "name": "App",
          "value": {{json .AppName}},
          "inline": true
        },{{end}}
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Name",
          "value": {{json .SSLName}},
          "inline": true
        },
        {
          "name": "Type",
          "value": {{json .SSLType}},
          "inline": true
        },
        {
          "name": "Domain",
          "value": {{json .Domain}},
          "inline": true
        },
        {
          "name": "Created At",
          "value": {{json .CreatedAt}},
          "inline": true
        },
        {
          "name": "Expire At",
          "value": {{json .ExpireAt}},
          "inline": true
        },
        {
          "name": "Expire In",
          "value": {{json .ExpireIn}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See object details",
          "value": {{json "[Go to Dashboard](" .DashboardLink ")"}},
          "inline": false
        }
      ]
    }
  ]
}