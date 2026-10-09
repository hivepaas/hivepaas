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
        "content": {{printf "%q" (print "[" .ProjectName "][" .AppName "] Deployment " (or (and .Succeeded "succeeded") "failed"))}}
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
              "content": {{printf "%q" (print "**Project:**\n" .ProjectName)}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**App:**\n" .AppName)}}
            }
          }{{if .Method | eq "repo"}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Repository:**\n" .RepoURL)}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Branch/Ref:**\n" .RepoRef)}}
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Commit Message:**\n" .CommitMsg)}}
            }
          },
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Commit Author:**\n" .CommitAuthor)}}
            }
          }{{else if .Method | eq "image"}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Image:**\n" .Image)}}
            }
          }{{end}}{{if .Reason}},
          {
            "is_short": false,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Reason:**\n" .Reason)}}
            }
          }{{end}},
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Started At:**\n" .StartedAt)}}
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": {{printf "%q" (print "**Duration:**\n" .Duration)}}
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
            "url": {{printf "%q" .DashboardLink}}
          }
        ]
      }
    ]
  }
}
