package bunex

import (
	"github.com/uptrace/bun"
)

var (
	selectNoneOption = func(query *bun.SelectQuery) *bun.SelectQuery {
		return query
	}
)

type SelectQueryOption func(*bun.SelectQuery) *bun.SelectQuery

func SelectColumns(cols ...string) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Column(cols...)
	}
}

func SelectColumnExpr(expr string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.ColumnExpr(expr, args...)
	}
}

func SelectExcludeColumns(cols ...string) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.ExcludeColumn(cols...)
	}
}

func SelectWithDeleted() SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.WhereAllWithDeleted()
	}
}

func SelectFor(selectFor string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.For(selectFor, args...)
	}
}

func SelectWhere(queryStr string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Where(queryStr, args...)
	}
}

func SelectWhereIf(cond bool, queryStr string, args ...any) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectWhere(queryStr, args...)
}

func SelectWhereOr(queryStr string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.WhereOr(queryStr, args...)
	}
}

func SelectWhereOrIf(cond bool, queryStr string, args ...any) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectWhereOr(queryStr, args...)
}

func SelectWhereGroup(opts ...SelectQueryOption) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.WhereGroup(" AND ", func(query *bun.SelectQuery) *bun.SelectQuery {
			for _, opt := range opts {
				query = opt(query)
			}
			return query
		})
	}
}

func SelectWhereOrGroup(opts ...SelectQueryOption) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.WhereGroup(" OR ", func(query *bun.SelectQuery) *bun.SelectQuery {
			for _, opt := range opts {
				query = opt(query)
			}
			return query
		})
	}
}

func SelectWhereIn[T any](queryStr string, slice ...T) SelectQueryOption {
	if len(slice) == 0 {
		return SelectWhere("1=0")
	}
	return SelectWhere(queryStr, In(slice))
}

func SelectWhereInIf[T any](cond bool, queryStr string, slice ...T) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectWhereIn(queryStr, slice...)
}

func SelectWhereNotIn[T any](queryStr string, slice ...T) SelectQueryOption {
	if len(slice) == 0 {
		return SelectWhere("1=1")
	}
	return SelectWhere(queryStr, In(slice))
}

func SelectWhereNotInIf[T any](cond bool, queryStr string, slice ...T) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectWhereNotIn(queryStr, slice...)
}

func SelectJoin(join string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Join(join, args...)
	}
}

func SelectJoinIf(cond bool, join string, args ...any) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectJoin(join, args...)
}

func SelectRelation(name string, opts ...SelectQueryOption) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Relation(name, func(relQry *bun.SelectQuery) *bun.SelectQuery {
			for _, opt := range opts {
				relQry = opt(relQry)
			}
			return relQry
		})
	}
}

func SelectRelationIf(cond bool, name string, opts ...SelectQueryOption) SelectQueryOption {
	if !cond {
		return selectNoneOption
	}
	return SelectRelation(name, opts...)
}

func SelectOrder(orderBy ...string) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Order(orderBy...)
	}
}

func SelectLimit(limit int) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Limit(limit)
	}
}

func SelectOffset(offset int) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Offset(offset)
	}
}

func SelectDistinct() SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Distinct()
	}
}

func SelectDistinctOn(queryStr string, args ...any) SelectQueryOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.DistinctOn(queryStr, args...)
	}
}

// ApplySelect applies extra select queries to the bun query
func ApplySelect(query *bun.SelectQuery, opts ...SelectQueryOption) *bun.SelectQuery {
	for _, opt := range opts {
		query = opt(query)
	}
	return query
}
