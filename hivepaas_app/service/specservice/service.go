// Package specservice exports HivePaaS configuration as a portable bundle.
//
// A spec is a faithful snapshot: it records values that are meaningful only on
// the installation that produced it, because the common case is re-import into
// that same installation. Portability is handled at import time by a
// skip-missing flag rather than by weakening what export records.
package specservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Export builds a configuration bundle for a scope. It reads only.
	Export(ctx context.Context, db database.IDB, req *ExportReq) (*ExportResp, error)
}
