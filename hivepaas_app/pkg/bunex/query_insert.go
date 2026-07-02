package bunex

import (
	"github.com/uptrace/bun"
)

type InsertQueryOption func(*bun.InsertQuery) *bun.InsertQuery

func InsertColumns(cols ...string) InsertQueryOption {
	return func(query *bun.InsertQuery) *bun.InsertQuery {
		return query.Column(cols...)
	}
}

func InsertValues(cols string, expr string, args ...any) InsertQueryOption {
	return func(query *bun.InsertQuery) *bun.InsertQuery {
		return query.Value(cols, expr, args...)
	}
}

func InsertExcludeColumns(cols ...string) InsertQueryOption {
	return func(query *bun.InsertQuery) *bun.InsertQuery {
		return query.ExcludeColumn(cols...)
	}
}

// ApplyInsert applies extra insert queries to the bun query
func ApplyInsert(query *bun.InsertQuery, opts ...InsertQueryOption) *bun.InsertQuery {
	for _, opt := range opts {
		query = opt(query)
	}
	return query
}
