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
        "content": "{{if .ProjectName | ne ""}}[{{.ProjectName}}]{{if .AppName | ne ""}}[{{.AppName}}]{{end}}{{else}}[System]{{end}} SSL renewal {{if .Succeeded}}succeeded{{else}}failed{{end}}"
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
              "content": "**Name:**\n{{.SSLName}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Type:**\n{{.SSLType}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Domain:**\n{{.Domain}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Created At:**\n{{.CreatedAt}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Expire At:**\n{{.ExpireAt}}"
            }
          }{{if .NextRenewalIn | gt 0}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Next Renewal In:**\n{{.NextRenewalIn}}"
            }
          }{{end}}
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
