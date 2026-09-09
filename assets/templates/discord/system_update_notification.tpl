{
  "embeds": [
    {
      "title": "System update {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {
          "name": "Current Version",
          "value": {{printf "%q" .CurrentVersion}},
          "inline": true
        },
        {
          "name": "Target Version",
          "value": {{printf "%q" .TargetVersion}},
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