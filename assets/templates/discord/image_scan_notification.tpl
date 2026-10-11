{
  "embeds": [
    {
      "title": {{json .Title}},
      "description": {{json .Lines}},
      "color": 15158332,
      "fields": [
        {
          "name": "See the images",
          "value": {{json "[Go to Dashboard](" .DashboardLink ")"}},
          "inline": false
        }
      ]
    }
  ]
}
