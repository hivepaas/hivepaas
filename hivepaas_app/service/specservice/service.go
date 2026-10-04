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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
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

	// ValidateImport reads an uploaded bundle and plans what importing it at a
	// scope would do. It writes nothing.
	ValidateImport(ctx context.Context, db database.IDB, req *ValidateImportReq) (*ValidateImportResp, error)

	// ApplyImport plans the request again on the caller's transaction and writes
	// what the plan says. It refuses a plan other than the one the operator saw,
	// one with a blocked issue, and one with issues nobody accepted.
	ApplyImport(ctx context.Context, db database.IDB, req *ApplyImportReq) (*ApplyImportResp, error)

	// PlanBundle plans a bundle built in memory - read from another format, such
	// as a compose file - as ValidateImport plans an uploaded one. It writes
	// nothing.
	PlanBundle(ctx context.Context, db database.IDB, req *PlanBundleReq) (*specmodel.ImportPlan, error)

	// ApplyBundle applies what PlanBundle planned, as ApplyImport applies what
	// ValidateImport did.
	ApplyBundle(ctx context.Context, db database.IDB, req *ApplyBundleReq) (*ApplyImportResp, error)
}
