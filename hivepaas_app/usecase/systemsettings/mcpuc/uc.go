package mcpuc

import (
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// enabledCacheTTL is how long the endpoint trusts what it last read of the
// switch. Every MCP request asks; a switch turned off takes effect within it, and
// at once on the backend that saved it.
const enabledCacheTTL = 10 * time.Second

type UC struct {
	*settings.BaseUC

	mu      sync.Mutex
	current entity.MCPSettings
	read    time.Time
}

func New(baseUC *settings.BaseUC) *UC {
	return &UC{BaseUC: baseUC}
}
