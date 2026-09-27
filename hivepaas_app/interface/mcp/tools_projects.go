package mcp

import (
	"context"
	"net/url"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// maxListed is the most items a lookup asks the API for.
const maxListed = 200

// Query parameters the API's lists and gets take.
const (
	paramPageLimit = "pageLimit"
	paramTrue      = "true"
)

// listProjects asks the project list endpoint, as the caller: each project
// with its envs. Names are looked up in it.
func listProjects(ctx context.Context, call *Call, search string) ([]*projectdto.ProjectResp, error) {
	query := url.Values{paramPageLimit: {strconv.Itoa(maxListed)}}
	if search != "" {
		query.Set(paramSearch, search)
	}
	var resp projectdto.ListProjectResp
	if err := call.Get(ctx, "/projects", query, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
