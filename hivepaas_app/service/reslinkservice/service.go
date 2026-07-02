package reslinkservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SetLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string, options ...SetLinkOption) error
	AddLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string, options ...SetLinkOption) error
	RemoveLinks(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstIDs []string) error
}
