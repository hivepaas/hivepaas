package bunex

import (
	"github.com/uptrace/bun"
)

var (
	updateNoneOption = func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query
	}
)

type UpdateQueryOption func(*bun.UpdateQuery) *bun.UpdateQuery

func UpdateColumns(cols ...string) UpdateQueryOption {
	return func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query.Column(cols...)
	}
}

func UpdateExcludeColumns(cols ...string) UpdateQueryOption {
	return func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query.ExcludeColumn(cols...)
	}
}

func UpdateWithDeleted() UpdateQueryOption {
	return func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query.WhereAllWithDeleted()
	}
}

func UpdateWhere(queryStr string, args ...any) UpdateQueryOption {
	return func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query.Where(queryStr, args...)
	}
}

func UpdateWhereIf(cond bool, queryStr string, args ...any) UpdateQueryOption {
	if !cond {
		return updateNoneOption
	}
	return UpdateWhere(queryStr, args...)
}

func UpdateWhereOr(queryStr string, args ...any) UpdateQueryOption {
	return func(query *bun.UpdateQuery) *bun.UpdateQuery {
		return query.WhereOr(queryStr, args...)
	}
}

func UpdateWhereOrIf(cond bool, queryStr string, args ...any) UpdateQueryOption {
	if !cond {
		return updateNoneOption
	}
	return UpdateWhereOr(queryStr, args...)
}

func UpdateWhereIn[T any](queryStr string, slice ...T) UpdateQueryOption {
	if len(slice) == 0 {
		return UpdateWhere("1=0")
	}
	return UpdateWhere(queryStr, List(slice))
}

func UpdateWhereInIf[T any](cond bool, queryStr string, slice ...T) UpdateQueryOption {
	if !cond {
		return updateNoneOption
	}
	return UpdateWhereIn(queryStr, slice...)
}

func UpdateWhereNotIn[T any](queryStr string, slice ...T) UpdateQueryOption {
	if len(slice) == 0 {
		return UpdateWhere("1=1")
	}
	return UpdateWhere(queryStr, List(slice))
}

func UpdateWhereNotInIf[T any](cond bool, queryStr string, slice ...T) UpdateQueryOption {
	if !cond {
		return updateNoneOption
	}
	return UpdateWhereNotIn(queryStr, slice...)
}

// ApplyUpdate applies extra update queries to the bun query
func ApplyUpdate(query *bun.UpdateQuery, opts ...UpdateQueryOption) *bun.UpdateQuery {
	for _, opt := range opts {
		query = opt(query)
	}
	return query
}
