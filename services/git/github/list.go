package github

import (
	gogithub "github.com/google/go-github/v85/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	defaultListPageSize = 100
	MaxListPageSize     = 100
)

func createListOpts(paging *basedto.Paging) (opts *gogithub.ListOptions, maxItems int) {
	opts = &gogithub.ListOptions{
		PerPage: defaultListPageSize,
	}
	if paging != nil {
		opts.Page = paging.ToPage()
		opts.PerPage = paging.ToPageSize()
	}
	maxItems = -1
	if opts.PerPage > MaxListPageSize {
		maxItems = opts.PerPage
		opts.PerPage = MaxListPageSize
	}
	return opts, maxItems
}
