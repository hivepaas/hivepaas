package bunex

import (
	"fmt"
	"strings"

	"github.com/uptrace/bun"
)

// ApplyUpsert updates inserting query to perform upserting (postgres only).
// Passing empty `conflictCols`, this func does nothing.
// Passing empty `updateCols`, this func performs `insert only` which skip conflicted data.
func ApplyUpsert(qry *bun.InsertQuery, conflictCols []string, updateCols []string) *bun.InsertQuery {
	if len(conflictCols) == 0 {
		return qry
	}

	action := "UPDATE"
	if len(updateCols) == 0 {
		action = "NOTHING"
	}
	if len(conflictCols) == 1 && strings.HasPrefix(conflictCols[0], "ON CONSTRAINT ") {
		qry = qry.On(fmt.Sprintf("CONFLICT %s DO %s", conflictCols[0], action))
	} else {
		qry = qry.On(fmt.Sprintf("CONFLICT (%s) DO %s", strings.Join(conflictCols, ","), action))
	}

	for _, col := range updateCols {
		qry = qry.Set(fmt.Sprintf("%s = EXCLUDED.%s", col, col))
	}
	return qry
}

// ApplyInsertIgnore inserts data with ignoring conflicted records
func ApplyInsertIgnore(qry *bun.InsertQuery, conflictCols []string) *bun.InsertQuery {
	return ApplyUpsert(qry, conflictCols, nil)
}
