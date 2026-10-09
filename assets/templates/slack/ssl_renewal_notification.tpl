{
  "attachments": [
    {
      "color": "{{if .Succeeded}}#2eb886{{else}}#a30200{{end}}",
      "title": {{json (scope .ProjectName .AppName) " SSL renewal " (outcome .Succeeded)}},
      "fields": [
        {{if .ProjectName | ne ""}}{
          "title": "Project",
          "value": {{json .ProjectName}},
          "short": true
        },{{end}}
        {{if .AppName | ne ""}}{
          "title": "App",
          "value": {{json .AppName}},
          "short": true
        },{{end}}
        {
          "title": "Name",
          "value": {{json .SSLName}},
          "short": true
        },
        {
          "title": "Type",
          "value": {{json .SSLType}},
          "short": true
        },
        {
          "title": "Domain",
          "value": {{json .Domain}},
          "short": true
        },
        {
          "title": "Created At",
          "value": {{json .CreatedAt}},
          "short": true
        },
        {
          "title": "Expire At",
          "value": {{json .ExpireAt}},
          "short": true
        },
        {{if .NextRenewalIn | gt 0}}{
          "title": "Next Renewal In",
          "value": {{json .NextRenewalIn}},
          "short": true
        },{{end}}{
          "title": "See task details",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}