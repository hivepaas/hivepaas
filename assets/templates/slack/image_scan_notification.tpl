{
  "attachments": [
    {
      "color": "#e74c3c",
      "title": {{json .Title}},
      "text": {{json .Lines}},
      "fields": [
        {
          "title": "See the images",
          "value": {{json "<" .DashboardLink "|Go to Dashboard>"}},
          "short": false
        }
      ],
      "mrkdwn_in": ["text", "fields"]
    }
  ]
}
