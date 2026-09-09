{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": "[{{.ProjectName}}][{{.AppName}}] Deployment {{if .Succeeded}}succeeded{{else}}failed{{end}}",
      "fields": [
        {
          "title": "Project",
          "value": {{printf "%q" .ProjectName}},
          "short": true
        },
        {
          "title": "App",
          "value": {{printf "%q" .AppName}},
          "short": true
        },
        {{if .Method | eq "repo"}}{
          "title": "Repository",
          "value": {{printf "%q" .RepoURL}},
          "short": true
        },
        {
          "title": "Branch/Ref",
          "value": {{printf "%q" .RepoRef}},
          "short": true
        },
        {
          "title": "Commit Message",
          "value": {{printf "%q" .CommitMsg}},
          "short": false
        },
        {
          "title": "Commit Author",
          "value": {{printf "%q" .CommitAuthor}},
          "short": false
        },
        {{else if .Method | eq "image"}}{
          "title": "Image",
          "value": {{printf "%q" .Image}},
          "short": false
        },
        {{end}}{
          "title": "Started At",
          "value": {{printf "%q" .StartedAt}},
          "short": true
        },
        {
          "title": "Duration",
          "value": {{printf "%q" .Duration}},
          "short": true
        },
        {
          "title": "See deployment details",
          "value": "<{{.DashboardLink}}|Go to Dashboard>",
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}
