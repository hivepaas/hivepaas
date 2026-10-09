{
  "embeds": [
    {
      "title": {{json (scope .ProjectName .AppName) " Healthcheck " (outcome .Succeeded)}},
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
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
          "value": {{json .HealthcheckName}},
          "inline": true
        },
        {
          "name": "Type",
          "value": {{json .HealthcheckType}},
          "inline": true
        },
        {
          "name": "Retries",
          "value": {{json .Retries}},
          "inline": true
        },
        {{if not .Succeeded}}{
          "name": "Expect",
          "value": {{json .Expect}},
          "inline": false
        },
        {
          "name": "Actual",
          "value": {{json .Actual}},
          "inline": false
        },{{end}}{
          "name": "Started At",
          "value": {{json .StartedAt}},
          "inline": true
        },
        {
          "name": "Duration",
          "value": {{json .Duration}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See task details",
          "value": {{json "[Go to Dashboard](" .DashboardLink ")"}},
          "inline": false
        }
      ]
    }
  ]
}
