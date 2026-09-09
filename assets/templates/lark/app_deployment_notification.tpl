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
        "content": "[{{.ProjectName}}][{{.AppName}}] Deployment {{if .Succeeded}}succeeded{{else}}failed{{end}}"
      }
    },
    "elements": [
      {
        "tag": "div",
        "fields": [
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Project:**\n{{.ProjectName}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**App:**\n{{.AppName}}"
            }
          }{{if .Method | eq "repo"}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Repository:**\n{{.RepoURL}}"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**Branch/Ref:**\n{{.RepoRef}}"
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": "**Commit Message:**\n{{.CommitMsg}}"
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": "**Commit Author:**\n{{.CommitAuthor}}"
            }
          }{{else if .Method | eq "image"}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": "**Image:**\n{{.Image}}"
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
