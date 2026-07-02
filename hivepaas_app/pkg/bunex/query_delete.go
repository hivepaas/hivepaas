package bunex

import (
	"github.com/uptrace/bun"
)

var (
	deleteNoneOption = func(query *bun.DeleteQuery) *bun.DeleteQuery {
		return query
	}
)

type DeleteQueryOption func(*bun.DeleteQuery) *bun.DeleteQuery

func DeleteWithDeleted() DeleteQueryOption {
	return func(query *bun.DeleteQuery) *bun.DeleteQuery {
		return query.WhereAllWithDeleted()
	}
}

func DeleteWhere(queryStr string, args ...any) DeleteQueryOption {
	return func(query *bun.DeleteQuery) *bun.DeleteQuery {
		return query.Where(queryStr, args...)
	}
}

func DeleteWhereIf(cond bool, queryStr string, args ...any) DeleteQueryOption {
	if !cond {
		return deleteNoneOption
	}
	return DeleteWhere(queryStr, args...)
}

func DeleteWhereOr(queryStr string, args ...any) DeleteQueryOption {
	return func(query *bun.DeleteQuery) *bun.DeleteQuery {
		return query.WhereOr(queryStr, args...)
	}
}

func DeleteWhereOrIf(cond bool, queryStr string, args ...any) DeleteQueryOption {
	if !cond {
		return deleteNoneOption
	}
	return DeleteWhereOr(queryStr, args...)
}

func DeleteWhereIn[T any](queryStr string, slice ...T) DeleteQueryOption {
	if len(slice) == 0 {
		return DeleteWhere("1=0")
	}
	return DeleteWhere(queryStr, List(slice))
}

func DeleteWhereInIf[T any](cond bool, queryStr string, slice ...T) DeleteQueryOption {
	if !cond {
		return deleteNoneOption
	}
	return DeleteWhereIn(queryStr, slice...)
}

func DeleteWhereNotIn[T any](queryStr string, slice ...T) DeleteQueryOption {
	if len(slice) == 0 {
		return DeleteWhere("1=1")
	}
	return DeleteWhere(queryStr, List(slice))
}

func DeleteWhereNotInIf[T any](cond bool, queryStr string, slice ...T) DeleteQueryOption {
	if !cond {
		return deleteNoneOption
	}
	return DeleteWhereNotIn(queryStr, slice...)
}

func DeleteWithForceDelete() DeleteQueryOption {
	return func(query *bun.DeleteQuery) *bun.DeleteQuery {
		return query.ForceDelete()
	}
}

func DeleteWithForceDeleteIf(cond bool) DeleteQueryOption {
	if !cond {
		return deleteNoneOption
	}
	return DeleteWithForceDelete()
}

// ApplyDelete applies extra delete queries to the bun query
func ApplyDelete(query *bun.DeleteQuery, opts ...DeleteQueryOption) *bun.DeleteQuery {
	for _, opt := range opts {
		query = opt(query)
	}
	return query
}
