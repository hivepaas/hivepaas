{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json "[" .ProjectName "][" .AppName "] Deployment " (outcome .Succeeded)}},
      "fields": [
        {
          "title": "Project",
          "value": {{json .ProjectName}},
          "short": true
        },
        {
          "title": "App",
          "value": {{json .AppName}},
          "short": true
        },
        {{if .Method | eq "repo"}}{
          "title": "Repository",
          "value": {{json .RepoURL}},
          "short": true
        },
        {
          "title": "Branch/Ref",
          "value": {{json .RepoRef}},
          "short": true
        },
        {
          "title": "Commit Message",
          "value": {{json .CommitMsg}},
          "short": false
        },
        {
          "title": "Commit Author",
          "value": {{json .CommitAuthor}},
          "short": false
        },
        {{else if .Method | eq "image"}}{
          "title": "Image",
          "value": {{json .Image}},
          "short": false
        },
        {{end}}{{if .Reason}}{
          "title": "Reason",
          "value": {{json .Reason}},
          "short": false
        },
        {{end}}{
          "title": "Started At",
          "value": {{json .StartedAt}},
          "short": true
        },
        {
          "title": "Duration",
          "value": {{json .Duration}},
          "short": true
        },
        {
          "title": "See deployment details",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}
