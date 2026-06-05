{
  "embeds": [
    {
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Scheduled task {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
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
          "name": "Scheduled Job",
          "value": {{printf "%q" .SchedJobName}},
          "inline": true
        },
        {
          "name": "Schedule",
          "value": {{printf "%q" .Schedule}},
          "inline": true
        },
        {
          "name": "Retries",
          "value": {{printf "%d" .Retries}},
          "inline": true
        },
        {
          "name": "Created At",
          "value": {{printf "%q" .CreatedAt}},
          "inline": true
        },
        {
          "name": "Started At",
          "value": {{printf "%q" .StartedAt}},
          "inline": true
        },
        {
          "name": "Duration",
          "value": {{printf "%q" .Duration}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See task details",
          "value": "[Go to Dashboard]({{.DashboardLink}})",
          "inline": false
        }
      ]
    }
  ]
}
