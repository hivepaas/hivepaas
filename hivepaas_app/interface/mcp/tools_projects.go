package mcp

import (
	"context"
	"net/url"
	"strconv"
)

// maxListed is the most items a list tool asks the API for.
const maxListed = 200

type listProjectsInput struct {
	Search string `json:"search,omitempty" jsonschema:"words in the project's key or name; empty for all"`
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

// apiProject is what the project endpoints answer, as far as the tools read it.
type apiProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Envs []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"envs"`
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
				item := projectItem{ID: p.ID, Key: p.Key, Name: p.Name, Envs: make([]string, 0, len(p.Envs))}
				for _, env := range p.Envs {
					item.Envs = append(item.Envs, env.Name)
				}
				out.Projects = append(out.Projects, item)
			}
			return out, nil
		})
}

// listProjects asks the project list endpoint, as the caller.
func listProjects(ctx context.Context, call *Call, search string) ([]apiProject, error) {
	query := url.Values{"pageLimit": {strconv.Itoa(maxListed)}}
	if search != "" {
		query.Set("search", search)
	}
	var resp struct {
		Data []apiProject `json:"data"`
	}
	if err := call.Get(ctx, "/projects", query, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
