{
  "msg_type": "interactive",
  "card": {
    "config": {
      "wide_screen_mode": true
    },
    "header": {
      "template": "yellow",
      "title": {
        "tag": "plain_text",
        "content": {{json (scope .ProjectName .AppName) " SSL expiring in " .ExpireIn}}
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
              "content": {{json "**Name:**\n" .SSLName}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Type:**\n" .SSLType}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Domain:**\n" .Domain}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Created At:**\n" .CreatedAt}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Expire At:**\n" .ExpireAt}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Expire In:**\n" .ExpireIn}}
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
