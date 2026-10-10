{
  "embeds": [
    {
      "title": {{json "[" .ProjectName "][" .AppName "] Clone " (outcome .Succeeded)}},
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {
          "name": "Project",
          "value": {{json .ProjectName}},
          "inline": true
        },
        {
          "name": "App",
          "value": {{json .AppName}},
          "inline": true
        },
        {"name": "​", "value": "​", "inline": true},
        {
          "name": "Clone",
          "value": {{json .CloneName " in " .CloneEnv}},
          "inline": false
        },
        {{if .Reason}}{
          "name": "Reason",
          "value": {{json .Reason}},
          "inline": false
        },
        {{end}}{
          "name": "Started At",
          "value": {{json .StartedAt}},
          "inline": true
        },
        {
          "name": "Duration",
          "value": {{json .Duration}},
          "inline": true
        },
        {"name": "​", "value": "​", "inline": true},
        {
          "name": "See clone details",
          "value": {{json "[Go to Dashboard](" .DashboardLink ")"}},
          "inline": false
        }
      ]
    }
  ]
}
