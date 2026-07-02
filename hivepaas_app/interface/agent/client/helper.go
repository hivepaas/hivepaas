package client

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

func CreateAuthCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", config.Current.Agent.SecretToken)
}

func CreateAuthCtxWithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(CreateAuthCtx(ctx))
}
