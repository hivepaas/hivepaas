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

	// BuildApp turns an AppDoc into the settings and service spec of an app being
	// provisioned. It persists nothing and creates no service: the caller does.
	// Only what specmodel.CheckBuildable accepts can be built.
	//
	// TODO: spec import - import builds each app of a bundle through this.
	BuildApp(ctx context.Context, db database.IDB, req *BuildAppReq) (*BuildAppResp, error)
}
