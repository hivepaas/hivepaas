{
  "embeds": [
    {
      "title": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL renewal {{if .Succeeded}}succeeded{{else}}failed{{end}}",
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
          "name": "Name",
          "value": {{printf "%q" .SSLName}},
          "inline": true
        },
        {
          "name": "Type",
          "value": {{printf "%q" .SSLType}},
          "inline": true
        },
        {
          "name": "Domain",
          "value": {{printf "%q" .Domain}},
          "inline": true
        },
        {
          "name": "Created At",
          "value": {{printf "%q" .CreatedAt}},
          "inline": true
        },
        {
          "name": "Expire At",
          "value": {{printf "%q" .ExpireAt}},
          "inline": true
        },
        {{if .NextRenewalIn | gt 0}}{
          "name": "Next Renewal In",
          "value": {{printf "%q" .NextRenewalIn}},
          "inline": true
        },{{end}}
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