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
        "content": {{json (scope .ProjectName .AppName) " Scheduled task " (outcome .Succeeded)}}
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
              "content": {{json "**Scheduled Job:**\n" .SchedJobName}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Schedule:**\n" .Schedule}}
            }
          },
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
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Retries:**\n" .Retries}}
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
