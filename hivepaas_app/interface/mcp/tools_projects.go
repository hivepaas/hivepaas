package mcp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// maxListed is the most items a list tool asks the API for.
const maxListed = 200

// Query parameters the API's lists and gets take.
const (
	paramPageLimit = "pageLimit"
	paramTrue      = "true"
)

type listProjectsInput struct {
	Search string `json:"search,omitempty" jsonschema:"text in the project's name or note; empty for all"`
}

type projectItem struct {
	ID   string   `json:"id"`
	Key  string   `json:"key"`
	Name string   `json:"name"`
	Envs []string `json:"envs"`
}

type projectList struct {
	Projects  []projectItem `json:"projects"`
	Truncated int           `json:"truncated,omitempty"`
}

func (l *projectList) shrink() bool {
	return shrinkList(&l.Projects, &l.Truncated)
}

func listProjectsTool() Tool {
	return readTool("list_projects", "List projects",
		"Lists the projects the API key's user can see, each with its envs. Start here to find the "+
			"project and env an app is in.",
		func(ctx context.Context, call *Call, in listProjectsInput) (projectList, error) {
			projects, err := listProjects(ctx, call, in.Search)
			if err != nil {
				return projectList{}, err
			}
			out := projectList{Projects: make([]projectItem, 0, len(projects))}
			for _, p := range projects {
				if p == nil {
					continue
				}
				item := projectItem{ID: p.ID, Key: p.Key, Name: p.Name, Envs: make([]string, 0, len(p.Envs))}
				for _, env := range p.Envs {
					if env != nil {
						item.Envs = append(item.Envs, env.Name)
					}
				}
				out.Projects = append(out.Projects, item)
			}
			return out, nil
		})
}

// listProjects asks the project list endpoint, as the caller: each project
// with its envs.
func listProjects(ctx context.Context, call *Call, search string) ([]*projectdto.ProjectResp, error) {
	query := url.Values{paramPageLimit: {strconv.Itoa(maxListed)}}
	if search != "" {
		query.Set("search", search)
	}
	var resp projectdto.ListProjectResp
	if err := call.Get(ctx, "/projects", query, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
