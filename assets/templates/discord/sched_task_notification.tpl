{
  "embeds": [
    {
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Scheduled task {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {{if .ProjectName | ne ""}}{
          "name": "Project",
          "value": {{printf "%q" .ProjectName}},
          "inline": false
        },{{end}}
        {{if .AppName | ne ""}}{
          "name": "App",
          "value": {{printf "%q" .AppName}},
          "inline": false
        },{{end}}
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Scheduled Job",
          "value": {{printf "%q" .SchedJobName}},
          "inline": false
        },
        {
          "name": "Schedule",
          "value": {{printf "%q" .Schedule}},
          "inline": false
        },
        {
          "name": "Started At",
          "value": {{printf "%q" .StartedAt}},
          "inline": false
        },
        {
          "name": "Duration",
          "value": {{printf "%q" .Duration}},
          "inline": false
        },
        {
          "name": "Retries",
          "value": "{{.Retries}}",
          "inline": false
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
