{
  "embeds": [
    {
      "title": {{json "[" .ProjectName "][" .AppName "] Deployment " (outcome .Succeeded)}},
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
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {{if .Method | eq "repo"}}{
          "name": "Repository",
          "value": {{json .RepoURL}},
          "inline": true
        },
        {
          "name": "Branch/Ref",
          "value": {{json .RepoRef}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Commit Message",
          "value": {{json .CommitMsg}},
          "inline": false
        },
        {
          "name": "Commit Author",
          "value": {{json .CommitAuthor}},
          "inline": false
        },
        {{else if .Method | eq "image"}}{
          "name": "Image",
          "value": {{json .Image}},
          "inline": false
        },
        {{end}}{{if .Reason}}{
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
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "See deployment details",
          "value": {{json "[Go to Dashboard](" .DashboardLink ")"}},
          "inline": false
        }
      ]
    }
  ]
}
