{
  "embeds": [
    {
      "title": "[{{.ProjectName}}][{{.AppName}}] Deployment {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "color": {{if .Succeeded}}3066993{{else}}15153724{{end}},
      "fields": [
        {
          "name": "Project",
          "value": {{printf "%q" .ProjectName}},
          "inline": true
        },
        {
          "name": "App",
          "value": {{printf "%q" .AppName}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {{if .Method | eq "repo"}}{
          "name": "Repository",
          "value": {{printf "%q" .RepoURL}},
          "inline": true
        },
        {
          "name": "Branch/Ref",
          "value": {{printf "%q" .RepoRef}},
          "inline": true
        },
        {"name": "\u200b", "value": "\u200b", "inline": true},
        {
          "name": "Commit Message",
          "value": {{printf "%q" .CommitMsg}},
          "inline": false
        },
        {
          "name": "Commit Author",
          "value": {{printf "%q" .CommitAuthor}},
          "inline": false
        },
        {{else if .Method | eq "image"}}{
          "name": "Image",
          "value": {{printf "%q" .Image}},
          "inline": false
        },
        {{end}}{
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
          "name": "See deployment details",
          "value": "[Go to Dashboard]({{.DashboardLink}})",
          "inline": false
        }
      ]
    }
  ]
}
