package bunex

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

// QueryRelation performs extra query relation(s) for the specific list of entities.
// Use this when relation can't be used in the main query.
func QueryRelation[T any, S ~[]T](ctx context.Context, db database.IDB, entities S,
	opts ...SelectQueryOption) (S, error) {
	var result S
	query := db.NewSelect().Model(&entities).WherePK()
	query = ApplySelect(query, opts...)

	err := query.Scan(ctx, &result)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return result, nil
}
