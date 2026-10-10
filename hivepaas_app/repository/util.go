package repository

import (
	"strings"

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
		return hperrors.NewArgumentInvalid("sort").WithCause(err)
	}
	return hperrors.Wrap(err)
}

// replaceNUL replaces each NUL byte with U+FFFD. A Postgres text column cannot
// hold one, and bun refuses to send it rather than drop it, failing the
// statement: text that comes from a process - its output, its error - goes
// through this before it is written.
func replaceNUL(s string) string {
	return strings.ReplaceAll(s, "\x00", "\uFFFD")
}

// inProjectEnvs keeps a project's query to the envs given, by the column that
// holds an env's ID; with none given, to all of them.
func inProjectEnvs(column string, envIDs []string) bunex.SelectQueryOption {
	return bunex.SelectWhereInIf(len(envIDs) > 0, column+" IN (?)", envIDs...)
}
