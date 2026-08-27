package repository

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func newPagingMeta(paging *basedto.Paging) *basedto.PagingMeta {
	if paging != nil {
		return &basedto.PagingMeta{
			Offset: paging.Offset,
			Limit:  paging.Limit,
		}
	}
	return &basedto.PagingMeta{}
}

func wrapPaginationError(err error, paging *basedto.Paging) error {
	if paging != nil && len(paging.Sort) > 0 && bunex.IsErrorColumnNotExist(err) {
		return hperrors.NewArgumentInvalidNT("sort").WithCause(err)
	}
	return hperrors.Wrap(err)
}
