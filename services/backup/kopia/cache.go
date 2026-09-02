package kopia

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// RefreshCache throws away what this client cached about the repository.
//
// The repository format - pack size among it - is cached locally on connect, and kopia has no
// command to refresh just that: `repository connect` over an existing config keeps the cached
// copy, and `cache sync` does not touch it either. Clearing the cache is what makes the next
// `repository status` read the repository itself. Verified against kopia, not assumed.
//
// The cost is the content and metadata cache going with it, so this belongs on an explicit,
// occasional read-back rather than on any hot path.
func (c *Client) RefreshCache(ctx context.Context) error {
	var errBuf bytes.Buffer
	_, err := c.execCommand(ctx, []string{cmdCache, "clear"}, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		return hperrors.Wrap(fmt.Errorf("%w: kopia cache clear: %s",
			backupmodel.ErrCommandFailed, strings.TrimSpace(errBuf.String())))
	}
	return nil
}
