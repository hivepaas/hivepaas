package bunex

import (
	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

func ApplyPagination(qry *bun.SelectQuery, paging *basedto.Paging) *bun.SelectQuery {
	if paging == nil {
		return qry
	}

	if paging.Offset > 0 {
		qry = qry.Offset(paging.Offset)
	}
	if paging.Limit > 0 {
		qry = qry.Limit(paging.Limit)
	}
	for _, order := range paging.Orders() {
		qry = qry.Order(order.ColumnName + " " + string(order.Direction))
	}
	return qry
}
