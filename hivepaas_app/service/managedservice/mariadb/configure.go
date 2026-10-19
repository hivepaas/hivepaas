package mariadb

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
	// serverArgs are what archiving needs of the server: its data directory a
	// directory of the volume, so that what HivePaaS keeps beside it is not
	// taken for a database; binary logs, row-based, with an id; and logs kept two
	// days, long enough that the server never drops one the archive loop has
	// not pushed.
	serverArgs = "mariadbd --datadir=/var/lib/mysql/data --log-bin=mariadb-bin --binlog-format=ROW " +
		"--server-id=1 --log-slave-updates --binlog-expire-logs-seconds=172800"
	// bufferPoolShare is the memory limit's share InnoDB caches pages in, a
	// quarter as Postgres's shared buffers, and minBufferPool MariaDB's default.
	bufferPoolShare = 4
	minBufferPool   = 128 * unit.MB
)

// Configure gives the server what archiving needs, and makes the engine's
// setting and jobs. Every database starts through walg-entrypoint, which runs
// the archive loop beside the server and hands over to the image's.
func (e *Engine) Configure(req *managedmodel.ConfigureReq) (*managedmodel.ConfigureResp, error) {
	options := walgkit.WithDefaults(req.Prepared.Options)
	source, err := walgkit.SetServerCommand(profile, req.Settings, serverCommand(req.Prepared.Doc))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := walgkit.Configure(profile, req, options, source)
	return resp, hperrors.Wrap(err)
}

func serverCommand(doc *specmodel.AppDoc) string {
	args := []string{serverArgs}
	if pool := bufferPool(doc); pool > 0 {
		args = append(args, fmt.Sprintf("--innodb-buffer-pool-size=%dM", pool/unit.MB))
	}
	return strings.Join(args, " ")
}

// bufferPool is a quarter of the memory limit, at least minBufferPool; zero
// without a limit, which leaves MariaDB its own default.
func bufferPool(doc *specmodel.AppDoc) unit.DataSize {
	if doc.Deployment == nil || doc.Deployment.Resources == nil || doc.Deployment.Resources.Limits == nil {
		return 0
	}
	limit := doc.Deployment.Resources.Limits.Memory
	if limit <= 0 {
		return 0
	}
	return max(limit/bufferPoolShare, minBufferPool)
}
