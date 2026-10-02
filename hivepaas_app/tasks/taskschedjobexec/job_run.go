package taskschedjobexec

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backuprepocleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslrenewalservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// jobRun is one scheduled job to run: the task data it runs under - its own
// task's, or a sequence's - its setting and the objects it references.
type jobRun struct {
	execData   *queue.TaskExecData
	jobSetting *entity.Setting
	refObjects *entity.RefObjects
	// sequence is set when the job runs as a step of a job sequence.
	sequence *schedjobexecservice.SequenceStep
}

// jobResult is what running a job says beyond its error. A container command
// gives its exit code, and, as a step of a sequence, its outputs, even when it
// fails.
type jobResult struct {
	skipNotification bool
	exitCode         *int
	outputs          map[string]string
}

// runJob runs a scheduled job by its type. A job run on its own schedule and a
// job run as a step of a sequence both come through here.
func (e *Executor) runJob(ctx context.Context, db database.Tx, run *jobRun) (*jobResult, error) {
	result := &jobResult{}
	schedJob := run.jobSetting.MustAsSchedJob()
	switch schedJob.JobType {
	case base.SchedJobTypeContainerCommand:
		resp, err := e.schedJobExecService.SchedJobExec(ctx, db, &schedjobexecservice.SchedJobExecReq{
			TaskExecData:    run.execData,
			SchedJobSetting: run.jobSetting,
			DestApp:         run.refObjects.RefApps[schedJob.App.ID],
			Sequence:        run.sequence,
		})
		if resp != nil {
			result.skipNotification = resp.SkipResultNotification
			result.exitCode = resp.ExitCode
			result.outputs = resp.Outputs
		}
		if err != nil {
			return result, hperrors.Wrap(err)
		}

	case base.SchedJobTypeSystemCleanup:
		setting := run.refObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return nil, hperrors.NewNotFound("System cleanup settings")
		}
		cleanupReq := &syscleanupservice.SysCleanupReq{
			TaskExecData:       run.execData,
			SysCleanupSettings: setting.MustAsSystemCleanup(),
		}
		cleanupReq.SetCleanupFlagsDefault()
		resp, err := e.sysCleanupService.Cleanup(ctx, db, cleanupReq)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		result.skipNotification = resp.SkipResultNotification

	case base.SchedJobTypeSystemBackup:
		setting := run.refObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return nil, hperrors.NewNotFound("System backup settings")
		}
		resp, err := e.sysBackupService.Backup(ctx, db, &sysbackupservice.SysBackupReq{
			TaskExecData:      run.execData,
			SysBackupSettings: setting.MustAsSystemBackup(),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		result.skipNotification = resp.SkipResultNotification

	case base.SchedJobTypeSSLRenewal:
		setting := run.refObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return nil, hperrors.NewNotFound("SSL renewal settings")
		}
		resp, err := e.sslRenewalService.SSLRenew(ctx, db, &sslrenewalservice.SSLRenewalReq{
			TaskExecData:      run.execData,
			RenewalJobSetting: run.jobSetting,
			RenewalSettings:   setting.MustAsSSLRenewal(),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		result.skipNotification = resp.SkipResultNotification

	case base.SchedJobTypeRegistryAuthRenewal:
		setting := run.refObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return nil, hperrors.NewNotFound("Registry auth renewal settings")
		}
		if err := e.renewRegistryAuths(ctx, db, run, setting.MustAsRegistryAuthRenewal()); err != nil {
			return nil, hperrors.Wrap(err)
		}

	case base.SchedJobTypeBackupRepoCleanup:
		setting := run.refObjects.RefSettings[schedJob.TargetSetting.ID]
		if setting == nil {
			return nil, hperrors.NewNotFound("Backup repo cleanup settings")
		}
		resp, err := e.backupRepoCleanupService.Cleanup(ctx, db, &backuprepocleanupservice.BackupRepoCleanupReq{
			TaskExecData:      run.execData,
			CleanupJobSetting: run.jobSetting,
			CleanupSettings:   setting.MustAsBackupRepoCleanup(),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		result.skipNotification = resp.SkipResultNotification

	case base.SchedJobTypeDataBackup:
		resp, err := e.dataBackupService.Backup(ctx, db, &databackupservice.BackupReq{
			TaskExecData: run.execData,
			JobSetting:   run.jobSetting,
			App:          run.refObjects.RefApps[schedJob.App.ID],
			RefObjects:   run.refObjects,
			Sequence:     run.sequence,
		})
		if err != nil {
			return result, hperrors.Wrap(err)
		}
		result.outputs = dataBackupOutputs(resp.Result)

	case base.SchedJobTypeFunctionInvoke:
		return e.invokeFunction(ctx, db, run, schedJob)

	case base.SchedJobTypeJobSequence:
		// Never a step: a sequence does not run another (checked on save).
		return nil, hperrors.NewUnsupported("A job sequence as a step")
	}

	return result, nil
}

// renewRegistryAuths runs a registry auth renewal: for the credentials the task
// names, or all of them. What it did is the task's output, failures included.
func (e *Executor) renewRegistryAuths(
	ctx context.Context,
	db database.Tx,
	run *jobRun,
	renewal *entity.RegistryAuthRenewal,
) error {
	task := run.execData.Task
	args, err := task.ArgsAsRegistryAuthRenewal()
	if err != nil {
		return hperrors.Wrap(err)
	}
	req := &registryauthservice.RenewReq{TaskExecData: run.execData, Interval: renewal.Interval()}
	if args != nil {
		req.TargetAuths = args.TargetAuths.ToIDStringSlice()
	}
	resp, err := e.registryAuthService.Renew(ctx, db, req)
	if resp != nil && resp.Output != nil {
		task.MustSetOutput(resp.Output)
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
