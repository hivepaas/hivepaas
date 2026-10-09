{
  "embeds": [
    {
      "title": {{json "System update " (outcome .Succeeded)}},
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {
          "name": "Current Version",
          "value": {{json .CurrentVersion}},
          "inline": true
        },
        {
          "name": "Target Version",
          "value": {{json .TargetVersion}},
          "inline": true
        },
        {
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