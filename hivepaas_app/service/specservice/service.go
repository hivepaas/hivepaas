// Package specservice exports HivePaaS configuration as a portable bundle.
//
// A spec is a faithful snapshot: it records values that are meaningful only on
// the installation that produced it, because the common case is re-import into
// that same installation. Portability is handled at import time by a
// skip-missing flag rather than by weakening what export records.
package specservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type Service interface {
	// Export builds a configuration bundle for a scope. It reads only.
	Export(ctx context.Context, db database.IDB, req *ExportReq) (*ExportResp, error)
}

type ExportReq struct {
	Scope       *entity.ObjectScope
	SecretsMode specmodel.SecretsMode
	// Passphrase is required when SecretsMode is encrypted.
	Passphrase string
	// WorkDir is where the bundle is staged and archived. The caller owns it and
	// is responsible for removing it once the response has been read.
	WorkDir string
}

type ExportResp struct {
	// Path is the finished archive on disk, inside WorkDir.
	Path string
	// Filename is what the download should be called.
	Filename string
	Size     int64
	Report   *specmodel.Report
}
