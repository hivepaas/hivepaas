package postgres

import (
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	// verifierScratchDir is where a verifier restores into, on a volume of its own.
	verifierScratchDir = "/var/lib/hivepaas-verify"

	// envRestoreSourceHost and envRestoreScratchDir tell walg-restore-test it runs
	// in a verifier: the live database it makes its restore point on is at the
	// host given - its app's key, on the project's network - and it restores into
	// the directory given, not beside the data.
	envRestoreSourceHost = "HIVEPAAS_RESTORE_SOURCE_HOST"
	envRestoreScratchDir = "HIVEPAAS_RESTORE_SCRATCH_DIR"
)

// Verifier idles: walg-entrypoint hands a command that is not postgres to the
// image's entrypoint, which runs it as it is, and walg-restore-test starts the
// server it restores into itself, with archiving off.
func (e *Engine) Verifier() managedmodel.VerifierSpec {
	return managedmodel.VerifierSpec{Command: "sleep infinity", ScratchDir: verifierScratchDir}
}

// moveRestoreTest points the restore test at the verifier given, or back at the
// database, and says whether its job changed. Its schedule stays.
func moveRestoreTest(app *entity.App, pg *entity.ManagedPostgres, setting *entity.Setting,
	change *managedmodel.RestoreTestChange) (bool, error) {
	if app == nil {
		return false, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("moving a restore test needs its app")
	}
	moved := entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer, CheckQuery: pg.RestoreTest.CheckQuery}
	jobType, target := base.SchedJobTypeContainerCommand, app.ID
	if change.Verifier != nil {
		moved.Mode = entity.ManagedRestoreTestVerifier
		moved.VerifierApp = entity.ObjectID{ID: change.Verifier.ID}
		moved.NodeID, moved.KeyAuth = change.NodeID, change.KeyAuth
		jobType, target = base.SchedJobTypeManagedRestoreTest, change.Verifier.ID
	}
	pg.RestoreTest = moved
	if setting == nil {
		return false, nil
	}

	job, err := setting.AsSchedJob()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	changed := job.JobType != jobType || job.App.ID != target
	job.JobType, job.App = jobType, entity.ObjectID{ID: target}
	setting.Kind = string(jobType)
	if change.Verifier != nil {
		changed = setCommandEnv(job.Command, envRestoreSourceHost, app.Key) || changed
		changed = setCommandEnv(job.Command, envRestoreScratchDir, verifierScratchDir) || changed
	} else if job.Command != nil {
		kept := slices.DeleteFunc(slices.Clone(job.Command.EnvVars), func(env *entity.EnvVar) bool {
			return env.Key == envRestoreSourceHost || env.Key == envRestoreScratchDir
		})
		changed = changed || len(kept) != len(job.Command.EnvVars)
		job.Command.EnvVars = kept
	}
	if !changed {
		return false, nil
	}
	if err = setting.SetData(job); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	return true, nil
}
