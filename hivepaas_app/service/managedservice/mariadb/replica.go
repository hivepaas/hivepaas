package mariadb

import (
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// What a member's server takes beside the command it shares with the cluster's
// app, which walg-entrypoint adds to it: its server id, unique in the cluster,
// and its binary logs' name. A member promoted keeps both: its logs go into the
// cluster's archive beside the others', under a name of their own.
const (
	envServerID = "HIVEPAAS_SERVER_ID"
	envLogBin   = "HIVEPAAS_LOG_BIN"
	// logBinPrefix is the cluster's app's logs' name, which a member's extends.
	logBinPrefix = "mariadb-bin"
	// firstMemberID is the first server id a member may have: the cluster's
	// app's is 1, in its command.
	firstMemberID = 2
)

// Replica is a standby the archive seeds, as every WAL-G engine's is
// (walgkit.Replica), with a server id and binary logs of its own.
func (e *Engine) Replica(req *managedmodel.ReplicaReq) (*managedmodel.ReplicaSpec, error) {
	spec, err := walgkit.Replica(profile, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	taken := map[int]bool{}
	for _, member := range req.Managed.Database.Members {
		taken[member.ServerID] = true
	}
	spec.ServerID = firstMemberID
	for taken[spec.ServerID] {
		spec.ServerID++
	}
	spec.Env = append(spec.Env, [2]string{envServerID, strconv.Itoa(spec.ServerID)},
		[2]string{envLogBin, logBinPrefix + "-" + req.Key})
	return spec, nil
}
