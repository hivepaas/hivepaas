{
  "msg_type": "interactive",
  "card": {
    "config": {
      "wide_screen_mode": true
    },
    "header": {
      "template": "{{if .Succeeded}}green{{else}}red{{end}}",
      "title": {
        "tag": "plain_text",
        "content": {{json (scope .ProjectName .AppName) " Healthcheck " (outcome .Succeeded)}}
      }
    },
    "elements": [
      {
        "tag": "div",
        "fields": [
          {{if .ProjectName | ne ""}}{
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Project:**\n" .ProjectName}}
            }
          },{{end}}
          {{if .AppName | ne ""}}{
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**App:**\n" .AppName}}
            }
          },{{end}}
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Name:**\n" .HealthcheckName}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Type:**\n" .HealthcheckType}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Retries:**\n" .Retries}}
            }
          }{{if not .Succeeded}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Expect:**\n" .Expect}}
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Actual:**\n" .Actual}}
            }
          }{{end}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Started At:**\n" .StartedAt}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Duration:**\n" .Duration}}
            }
          }
        ]
      },
      {
        "tag": "action",
        "actions": [
          {
            "tag": "button",
            "text": {
              "tag": "plain_text",
              "content": "Go to Dashboard"
            },
            "type": "primary",
            "url": {{json .DashboardLink}}
          }
        ]
      }
    ]
  }
}
