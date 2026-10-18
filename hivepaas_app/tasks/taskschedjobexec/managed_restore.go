package taskschedjobexec

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

const (
	// verifierTaskFindRetryMax and verifierTaskFindRetryDelay wait for the
	// verifier's task: scaled up from none, it is scheduled, its image pulled on
	// a node that may not have it yet, and started - two minutes at most.
	verifierTaskFindRetryMax   = 24
	verifierTaskFindRetryDelay = 5 * time.Second
	verifierScaleRetries       = 2
)

// runRestoreTest runs a managed database's restore test in its verifier app.
// The verifier is up only for its test: scaled to one, the script run in its
// task once that runs, and scaled back to none however the test ends - a
// verifier left up holds a copy of the database, on a node that may be needed.
// No license is checked: a verifier made under a plan that has since lapsed
// still tests what it was made to.
func (e *Executor) runRestoreTest(
	ctx context.Context,
	db database.Tx,
	run *jobRun,
	job *entity.SchedJob,
) (*jobResult, error) {
	verifier := run.refObjects.RefApps[job.App.ID]
	if verifier == nil {
		return nil, hperrors.NewNotFound("Verifier app")
	}
	if err := e.scaleVerifier(ctx, verifier, 1); err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() {
		// Without the run's cancellation: a canceled run scales it down too.
		ctx := context.WithoutCancel(ctx)
		if err := e.scaleVerifier(ctx, verifier, 0); err != nil {
			_ = run.execData.LogStore.Add(ctx, tasklog.NewErrFrame(
				"The verifier app could not be scaled down: "+err.Error(), tasklog.TsNow))
		}
	}()

	resp, err := e.schedJobExecService.SchedJobExec(ctx, db, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:       run.execData,
		SchedJobSetting:    run.jobSetting,
		DestApp:            verifier,
		TaskFindRetryMax:   verifierTaskFindRetryMax,
		TaskFindRetryDelay: verifierTaskFindRetryDelay,
		Sequence:           run.sequence,
	})
	result := &jobResult{}
	if resp != nil {
		result.skipNotification = resp.SkipResultNotification
		result.exitCode = resp.ExitCode
		result.outputs = resp.Outputs
	}
	if err != nil {
		return result, hperrors.Wrap(err)
	}
	return result, nil
}

// scaleVerifier sets the verifier's replicas.
func (e *Executor) scaleVerifier(ctx context.Context, verifier *entity.App, replicas uint64) error {
	err := e.dockerManager.ServiceUpdateFunc(ctx, verifier.ServiceID, nil,
		func(_ int, svc *swarm.Service) (bool, error) {
			if svc.Spec.Mode.Replicated == nil {
				return false, hperrors.NewUnsupported("A verifier app that is not replicated")
			}
			if svc.Spec.Mode.Replicated.Replicas != nil && *svc.Spec.Mode.Replicated.Replicas == replicas {
				return false, nil
			}
			svc.Spec.Mode.Replicated.Replicas = &replicas
			return true, nil
		}, verifierScaleRetries, 0)
	return hperrors.Wrap(err)
}
