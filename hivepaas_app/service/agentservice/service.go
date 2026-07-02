package agentservice

import (
	"context"
)

type Service interface {
	GetAgentAddrForNode(ctx context.Context, nodeID string) (string, error)
}
