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
        "content": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} Healthcheck {{if .Succeeded}}succeeded{{else}}failed{{end}}"
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
              "content": "**Project:**\n{{.ProjectName}}"
            }
          },{{end}}
          {{if .AppName | ne ""}}{
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**App:**\n{{.AppName}}"
            }
          },{{end}}
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Name:**\n{{.HealthcheckName}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Type:**\n{{.HealthcheckType}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Retries:**\n{{.Retries}}"
            }
          }{{if not .Succeeded}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": "**Expect:**\n{{.Expect}}"
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": "**Actual:**\n{{.Actual}}"
            }
          }{{end}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Started At:**\n{{.StartedAt}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Duration:**\n{{.Duration}}"
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
            "url": "{{.DashboardLink}}"
          }
        ]
      }
    ]
  }
}
