package walgkit

import (
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	// PromoteScript promotes the standby it runs in; a template that ships it
	// ships walg-entrypoint's fenced mode and promoted marker too.
	PromoteScript = "walg-promote"
	// EnvFenced keeps a database from starting its server: walg-entrypoint idles.
	EnvFenced = "HIVEPAAS_FENCED"
)

// Promote fences what writes now and promotes a replica (design §9, "Manual
// promotion"): the old primary is stopped and kept from starting, so no two
// primaries write to the cluster's archive. A member sheds the standby's
// variables, and the other replicas are pointed at it by the host's.
func Promote(p *Profile, req *managedmodel.PromoteReq) (*managedmodel.PromoteSpec, error) {
	managed := req.Managed
	if managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	db := managed.Database
	if db == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if db.Restore != nil && db.Restore.ReadyAt.IsZero() {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	if req.Member == nil || req.Member.Role != entity.ManagedRoleReplica {
		return nil, hperrors.NewArgumentInvalid("member").WithExtraDetail("only a replica is promoted")
	}
	promotion := p.Promotion
	if promotion.Command == "" {
		promotion = Promotion{Command: ScriptDir + "/" + PromoteScript, Requires: PromoteScript}
	}
	if !slices.Contains(req.MountPaths, ScriptDir+"/"+promotion.Requires) {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedPromoteUnsupported).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	return &managedmodel.PromoteSpec{
		FenceEnv:       [][2]string{{EnvFenced, req.Member.Key + " was promoted"}},
		Promote:        &entity.CommandTemplate{Command: promotion.Command},
		Allow:          &entity.CommandTemplate{Command: ScriptDir + "/" + ReplicationScript},
		StandbyEnvKeys: []string{EnvStandbyHost, EnvStandbyName, EnvStandbyPasswordFile},
		StandbyHostKey: EnvStandbyHost,
		Alias:          db.Alias,
	}, nil
}

// promoteWriter makes the member named the cluster's primary: what wrote before
// is fenced - the app itself, or the member promoted before - and the engine's
// jobs move to it.
func promoteWriter(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	db, change := req.Managed.Database, req.Change
	if change.Writer == nil {
		return nil
	}
	member := db.Member(*change.Writer)
	if member == nil || member.Role != entity.ManagedRoleReplica {
		return hperrors.NewArgumentInvalid("member").
			WithExtraDetail("%s is not a replica of this database", *change.Writer)
	}
	if old := db.Member(db.Writer.ID); old != nil {
		old.Role = entity.ManagedRoleFenced
	} else if db.FencedAt.IsZero() {
		db.FencedAt = req.Now
	}
	member.Role = entity.ManagedRolePrimary
	db.Writer = member.App
	if db.BackupFrom.ID == member.App.ID {
		db.BackupFrom = entity.ObjectID{}
	}
	return pointJobsAtWriter(req, resp)
}

// pointJobsAtWriter runs the engine's jobs on the cluster's primary: the nightly
// on the replica its backups are taken on, if any; the restore test in a
// verifier makes its point on the primary.
func pointJobsAtWriter(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	db := req.Managed.Database
	writer := gofn.Coalesce(db.Writer.ID, req.App.ID)
	targets := map[string]string{
		db.Jobs.Nightly.ID: gofn.Coalesce(db.BackupFrom.ID, writer),
		db.Jobs.Health.ID:  writer,
		db.Jobs.Dump.ID:    writer,
	}
	if db.RestoreTest.Mode != entity.ManagedRestoreTestVerifier {
		targets[db.Jobs.RestoreTest.ID] = writer
	}
	for id, target := range targets {
		setting := req.Jobs[id]
		if id == "" || setting == nil {
			continue
		}
		job, err := setting.AsSchedJob()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if job.App.ID == target {
			continue
		}
		job.App = entity.ObjectID{ID: target}
		if err = changeJob(setting, job, resp); err != nil {
			return err
		}
	}
	if setting := req.Jobs[db.Jobs.RestoreTest.ID]; setting != nil &&
		db.RestoreTest.Mode == entity.ManagedRestoreTestVerifier {
		job, err := setting.AsSchedJob()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if SetCommandEnv(job.Command, EnvRestoreSourceHost, db.WriterKey(req.App)) {
			if err = changeJob(setting, job, resp); err != nil {
				return err
			}
		}
	}
	return nil
}

func changeJob(setting *entity.Setting, job *entity.SchedJob, resp *managedmodel.ReconfigureResp) error {
	if err := setting.SetData(job); err != nil {
		return hperrors.Wrap(err)
	}
	if !slices.Contains(resp.Changed, setting) {
		setting.UpdateVer++
		resp.Changed = append(resp.Changed, setting)
	}
	return nil
}
