package postgres

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

const (
	// replicationScript lets replicas into the primary: its password on stdin.
	// A template that ships it ships walg-entrypoint's standby mode too.
	replicationScript = "walg-replication"
	// replicaPasswordFile is where a replica finds the replication password, from
	// replicaSecretName, readable by the postgres user alone.
	replicaPasswordFile = "/run/secrets/hivepaas-replication/password" //nolint:gosec // G101: a path
	replicaSecretName   = "HIVEPAAS_REPLICATION_PASSWORD"              //nolint:gosec // G101: a secret's name

	// What walg-entrypoint's standby mode reads: the primary's host, what the
	// primary knows the replica by, and the password's file.
	envStandbyHost         = "HIVEPAAS_STANDBY_HOST"
	envStandbyName         = "HIVEPAAS_STANDBY_NAME"
	envStandbyPasswordFile = "HIVEPAAS_STANDBY_PASSWORD_FILE" //nolint:gosec // G101: a variable's name

	// outReplicas is walg-health's list of the replicas streaming from the
	// primary: space-separated name@state@bytes@seconds.
	outReplicas = "REPLICAS"
)

// Replica is a standby the archive seeds, on a node of its own: walg-entrypoint
// fetches the latest base backup and streams from the primary, by its key on the
// project's network. Only a database whose scripts can, archiving off-site and
// ready, has one; and only a primary.
func (e *Engine) Replica(req *managedmodel.ReplicaReq) (*managedmodel.ReplicaSpec, error) {
	managed := req.Managed
	if managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	pg := managed.Database
	if pg == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if !slices.Contains(req.MountPaths, walgkit.ScriptDir+"/"+replicationScript) {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedReplicaUnsupported).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	if pg.Restore != nil && pg.Restore.ReadyAt.IsZero() {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	// A volume repository's directory could be shared; for now a replica is
	// seeded from object storage only, as the verifier is.
	if pg.Storage.Placement != entity.ManagedPlacementOffSite {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedReplicaNeedsObjectStorage)
	}
	target := ""
	for _, mount := range managed.Locks.Mounts {
		if mount != walgkit.VolumeArchiveDir {
			target = mount
			break
		}
	}
	if target == "" {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("postgres: the primary's data mount is not known")
	}
	writer := req.Primary
	if req.Writer != nil {
		writer = req.Writer
	}
	return &managedmodel.ReplicaSpec{
		Env: [][2]string{
			{envStandbyHost, writer.Key},
			{envStandbyName, req.Key},
			{envStandbyPasswordFile, replicaPasswordFile},
		},
		SecretName:   replicaSecretName,
		PasswordFile: replicaPasswordFile,
		DataTarget:   target,
		Allow:        &entity.CommandTemplate{Command: walgkit.ScriptDir + "/" + replicationScript},
	}, nil
}

// membersStatus is each member as the primary's last health check saw it.
func membersStatus(members []*entity.ManagedMember, health *managedmodel.JobRun) []*managedmodel.MemberStatus {
	if len(members) == 0 {
		return nil
	}
	var seen map[string]*managedmodel.MemberStatus
	var checkedAt *time.Time
	if health != nil {
		if list, listed := health.Outputs[outReplicas]; listed {
			seen = replicasOf(list)
			checkedAt = &health.EndedAt
		}
	}
	out := make([]*managedmodel.MemberStatus, 0, len(members))
	for _, member := range members {
		status := &managedmodel.MemberStatus{Member: member, State: managedmodel.MemberUnknown}
		// The primary and a fenced member stream from none.
		if member.Role == entity.ManagedRolePrimary || member.Role == entity.ManagedRoleFenced {
			status.State = member.Role
			out = append(out, status)
			continue
		}
		if seen != nil {
			status.State, status.CheckedAt = managedmodel.MemberDisconnected, checkedAt
			if one := seen[member.Key]; one != nil {
				status.State, status.LagBytes, status.LagSeconds = one.State, one.LagBytes, one.LagSeconds
			}
		}
		out = append(out, status)
	}
	return out
}

// replicasOf reads outReplicas, by name.
func replicasOf(value string) map[string]*managedmodel.MemberStatus {
	out := map[string]*managedmodel.MemberStatus{}
	for _, field := range strings.Fields(value) {
		parts := strings.Split(field, "@")
		if len(parts) != 4 { //nolint:mnd // name@state@bytes@seconds
			continue
		}
		status := &managedmodel.MemberStatus{State: parts[1]}
		status.LagBytes, _ = strconv.ParseInt(parts[2], 10, 64)
		if seconds, err := strconv.ParseFloat(parts[3], 64); err == nil {
			status.LagSeconds = &seconds
		}
		out[parts[0]] = status
	}
	return out
}
