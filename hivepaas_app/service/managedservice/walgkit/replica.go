package walgkit

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	// ReplicationScript lets replicas into the member that writes: the
	// replication password on stdin. A template that ships it ships
	// walg-entrypoint's standby mode too.
	ReplicationScript = "walg-replication"
	// ReplicaPasswordFile is where a replica finds the replication password, from
	// ReplicaSecretName, readable by the engine's file owner alone.
	ReplicaPasswordFile = "/run/secrets/hivepaas-replication/password" //nolint:gosec // G101: a path
	ReplicaSecretName   = "HIVEPAAS_REPLICATION_PASSWORD"              //nolint:gosec // G101: a secret's name

	// What walg-entrypoint's standby mode reads: the writer's host, what the
	// writer knows the replica by, and the password's file.
	EnvStandbyHost         = "HIVEPAAS_STANDBY_HOST"
	EnvStandbyName         = "HIVEPAAS_STANDBY_NAME"
	EnvStandbyPasswordFile = "HIVEPAAS_STANDBY_PASSWORD_FILE" //nolint:gosec // G101: a variable's name

	// OutReplicas is walg-health's list of the replicas streaming from the
	// member that writes: space-separated name@state@bytes@seconds.
	OutReplicas = "REPLICAS"
)

// Replica is a standby the archive seeds, on a node of its own: walg-entrypoint
// fetches the latest base backup and streams from the cluster's writer, by its
// key on the project's network. Only a database whose scripts can, archiving
// off-site and ready, has one; and only a cluster's app. The engine adds what
// its own server needs.
func Replica(p *Profile, req *managedmodel.ReplicaReq) (*managedmodel.ReplicaSpec, error) {
	managed := req.Managed
	if managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	db := managed.Database
	if db == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if !slices.Contains(req.MountPaths, ScriptDir+"/"+ReplicationScript) {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedReplicaUnsupported).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	if db.Restore != nil && db.Restore.ReadyAt.IsZero() {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	// A volume repository's directory could be shared; for now a replica is
	// seeded from object storage only, as the verifier is.
	if db.Storage.Placement != entity.ManagedPlacementOffSite {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedReplicaNeedsObjectStorage)
	}
	target := ""
	for _, mount := range managed.Locks.Mounts {
		if mount != VolumeArchiveDir {
			target = mount
			break
		}
	}
	if target == "" {
		return nil, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("%s: the primary's data mount is not known", p.Engine)
	}
	writer := req.Primary
	if req.Writer != nil {
		writer = req.Writer
	}
	return &managedmodel.ReplicaSpec{
		Env: [][2]string{
			{EnvStandbyHost, writer.Key},
			{EnvStandbyName, req.Key},
			{EnvStandbyPasswordFile, ReplicaPasswordFile},
		},
		SecretName:   ReplicaSecretName,
		PasswordFile: ReplicaPasswordFile,
		FileOwner:    p.FileOwner,
		DataTarget:   target,
		Allow:        &entity.CommandTemplate{Command: ScriptDir + "/" + ReplicationScript},
	}, nil
}

// MembersStatus is each member as the writer's last health check saw it.
func MembersStatus(members []*entity.ManagedMember, health *managedmodel.JobRun) []*managedmodel.MemberStatus {
	if len(members) == 0 {
		return nil
	}
	var seen map[string]*managedmodel.MemberStatus
	var checkedAt *time.Time
	if health != nil {
		if list, listed := health.Outputs[OutReplicas]; listed {
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

// replicasOf reads OutReplicas, by name.
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
