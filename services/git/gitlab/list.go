package gitlab

import (
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	defaultListPageSize = 100
	MaxListPageSize     = 100
)

func createListOpts(paging *basedto.Paging) (opts *gogitlab.ListOptions, maxItems int64) {
	opts = &gogitlab.ListOptions{
		PerPage: defaultListPageSize,
	}
	if paging != nil {
		opts.Page = int64(paging.ToPage())
		opts.PerPage = int64(paging.ToPageSize())
	}
	maxItems = -1
	if opts.PerPage > MaxListPageSize {
		maxItems = opts.PerPage
		opts.PerPage = MaxListPageSize
	}
	return opts, maxItems
}
