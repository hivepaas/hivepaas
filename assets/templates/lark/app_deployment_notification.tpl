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
        "content": {{json "[" .ProjectName "][" .AppName "] Deployment " (outcome .Succeeded)}}
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
              "content": {{json "**Project:**\n" .ProjectName}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**App:**\n" .AppName}}
            }
          }{{if .Method | eq "repo"}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Repository:**\n" .RepoURL}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Branch/Ref:**\n" .RepoRef}}
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Commit Message:**\n" .CommitMsg}}
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Commit Author:**\n" .CommitAuthor}}
            }
          }{{else if .Method | eq "image"}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Image:**\n" .Image}}
            }
          }{{end}}{{if .Reason}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{json "**Reason:**\n" .Reason}}
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
