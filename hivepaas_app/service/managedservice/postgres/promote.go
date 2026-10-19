package postgres

import (
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

const (
	// promoteScript promotes the standby it runs in; a template that ships it
	// ships walg-entrypoint's fenced mode and promoted marker too.
	promoteScript = "walg-promote"
	// envFenced keeps a database from starting its server: walg-entrypoint idles.
	envFenced = "HIVEPAAS_FENCED"
)

// Promote fences what writes now and promotes a replica (design §9, "Manual
// promotion"): the old primary is stopped and kept from starting, so no two
// primaries write to the cluster's archive.
func (e *Engine) Promote(req *managedmodel.PromoteReq) (*managedmodel.PromoteSpec, error) {
	managed := req.Managed
	if managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	pg := managed.Database
	if pg == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if pg.Restore != nil && pg.Restore.ReadyAt.IsZero() {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSwitchNotReady).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	if req.Member == nil || req.Member.Role != entity.ManagedRoleReplica {
		return nil, hperrors.NewArgumentInvalid("member").WithExtraDetail("only a replica is promoted")
	}
	if !slices.Contains(req.MountPaths, walgkit.ScriptDir+"/"+promoteScript) {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedPromoteUnsupported).
			WithParam("App", req.Primary.Name).ExposeParams("App")
	}
	return &managedmodel.PromoteSpec{
		FenceEnv:       [][2]string{{envFenced, req.Member.Key + " was promoted"}},
		Promote:        &entity.CommandTemplate{Command: walgkit.ScriptDir + "/" + promoteScript},
		Allow:          &entity.CommandTemplate{Command: walgkit.ScriptDir + "/" + replicationScript},
		StandbyEnvKeys: []string{envStandbyHost, envStandbyName, envStandbyPasswordFile},
		StandbyHostKey: envStandbyHost,
	}, nil
}

// promoteWriter makes the member named the cluster's primary: what wrote before
// is fenced - the app itself, or the member promoted before - and the engine's
// jobs move to it.
func promoteWriter(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	pg, change := req.Managed.Database, req.Change
	if change.Writer == nil {
		return nil
	}
	member := pg.Member(*change.Writer)
	if member == nil || member.Role != entity.ManagedRoleReplica {
		return hperrors.NewArgumentInvalid("member").
			WithExtraDetail("%s is not a replica of this database", *change.Writer)
	}
	if old := pg.Member(pg.Writer.ID); old != nil {
		old.Role = entity.ManagedRoleFenced
	} else if pg.FencedAt.IsZero() {
		pg.FencedAt = req.Now
	}
	member.Role = entity.ManagedRolePrimary
	pg.Writer = member.App
	if pg.BackupFrom.ID == member.App.ID {
		pg.BackupFrom = entity.ObjectID{}
	}
	return pointJobsAtWriter(req, resp)
}

// pointJobsAtWriter runs the engine's jobs on the cluster's primary: the nightly
// on the replica its backups are taken on, if any; the restore test in a
// verifier makes its point on the primary.
func pointJobsAtWriter(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	pg := req.Managed.Database
	writer := gofn.Coalesce(pg.Writer.ID, req.App.ID)
	targets := map[string]string{
		pg.Jobs.Nightly.ID: gofn.Coalesce(pg.BackupFrom.ID, writer),
		pg.Jobs.Health.ID:  writer,
		pg.Jobs.Dump.ID:    writer,
	}
	if pg.RestoreTest.Mode != entity.ManagedRestoreTestVerifier {
		targets[pg.Jobs.RestoreTest.ID] = writer
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
	if setting := req.Jobs[pg.Jobs.RestoreTest.ID]; setting != nil &&
		pg.RestoreTest.Mode == entity.ManagedRestoreTestVerifier {
		job, err := setting.AsSchedJob()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if walgkit.SetCommandEnv(job.Command, walgkit.EnvRestoreSourceHost, pg.WriterKey(req.App)) {
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
