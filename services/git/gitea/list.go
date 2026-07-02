package gitea

import (
	gogitea "code.gitea.io/sdk/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	defaultListPageSize = 100
	MaxListPageSize     = 100
)

func createListOpts(paging *basedto.Paging) (opts *gogitea.ListOptions, maxItems int) {
	opts = &gogitea.ListOptions{
		PageSize: defaultListPageSize,
	}
	if paging != nil {
		opts.Page = paging.ToPage()
		opts.PageSize = paging.ToPageSize()
	}
	maxItems = -1
	if opts.PageSize > MaxListPageSize {
		maxItems = opts.PageSize
		opts.PageSize = MaxListPageSize
	}
	return opts, maxItems
}
