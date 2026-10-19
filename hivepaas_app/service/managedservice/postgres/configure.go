package postgres

import (
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	// sharedBuffersShare is the memory limit's share Postgres caches pages in,
	// the usual quarter, and minSharedBuffers its floor.
	sharedBuffersShare = 4
	minSharedBuffers   = 128 * unit.MB
)

// Configure gives the server what archiving needs, and makes the engine's
// setting and jobs. Every database starts through walg-entrypoint: it hands
// over to the image's, with a checkpoint once the server is up so that
// archive_timeout holds from the start, and recovers a restored database first.
func (e *Engine) Configure(req *managedmodel.ConfigureReq) (*managedmodel.ConfigureResp, error) {
	options := walgkit.WithDefaults(req.Prepared.Options)
	source, err := walgkit.SetServerCommand(profile, req.Settings, serverCommand(req.Prepared.Doc, options))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := walgkit.Configure(profile, req, options, source)
	return resp, hperrors.Wrap(err)
}

// serverCommand is the server's: archiving on, every WAL segment pushed by
// walg-run, and a share of the memory limit for shared buffers.
func serverCommand(doc *specmodel.AppDoc, options managedmodel.Options) string {
	args := []string{
		"postgres",
		"-c wal_level=replica",
		"-c archive_mode=on",
		fmt.Sprintf("-c archive_timeout=%ds", int(options.ArchiveTimeout.Seconds())),
		`-c "archive_command=` + walgkit.ScriptDir + `/walg-run wal-push %p"`,
	}
	if buffers := sharedBuffers(doc); buffers > 0 {
		args = append(args, fmt.Sprintf("-c shared_buffers=%dMB", buffers/unit.MB))
	}
	return strings.Join(args, " ")
}

// sharedBuffers is a quarter of the memory limit, at least minSharedBuffers;
// zero without a limit, which leaves Postgres its own default.
func sharedBuffers(doc *specmodel.AppDoc) unit.DataSize {
	if doc.Deployment == nil || doc.Deployment.Resources == nil || doc.Deployment.Resources.Limits == nil {
		return 0
	}
	limit := doc.Deployment.Resources.Limits.Memory
	if limit <= 0 {
		return 0
	}
	return max(limit/sharedBuffersShare, minSharedBuffers)
}
