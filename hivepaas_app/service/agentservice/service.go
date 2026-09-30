package agentservice

import (
	"context"
)

type Service interface {
	GetAgentAddrForNode(ctx context.Context, nodeID string) (string, error)
	GetAgentAddrForNodeLabel(ctx context.Context, nodeLabel string) (string, error)
	// NodeIDsWithLabel is every ready node carrying the label, key or key=value.
	NodeIDsWithLabel(ctx context.Context, nodeLabel string) ([]string, error)
}
