{
  "embeds": [
    {
      "title": {{json (scope .ProjectName .AppName) " Scheduled task " (outcome .Succeeded)}},
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {{if .ProjectName | ne ""}}{
          "name": "Project",
          "value": {{json .ProjectName}},
          "inline": false
        },{{end}}
        {{if .AppName | ne ""}}{
          "name": "App",
          "value": {{json .AppName}},
          "inline": false
        },{{end}}
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Scheduled Job",
          "value": {{json .SchedJobName}},
          "inline": false
        },
        {
          "name": "Schedule",
          "value": {{json .Schedule}},
          "inline": false
        },
        {
          "name": "Started At",
          "value": {{json .StartedAt}},
          "inline": false
        },
        {
          "name": "Duration",
          "value": {{json .Duration}},
          "inline": false
        },
        {
          "name": "Retries",
          "value": {{json .Retries}},
          "inline": false
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
