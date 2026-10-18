package taskschedjobexec

import (
	"context"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
)

// reconcileArtifacts has every app's artifacts on the nodes its tasks may start
// on. Its log names each app that did not, with why; the run then fails, and
// notifies no one: the next is minutes away.
func (e *Executor) reconcileArtifacts(ctx context.Context, db database.Tx, run *jobRun) (*jobResult, error) {
	resp, err := e.artifactService.ReconcileAll(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_ = run.execData.LogStore.Add(ctx, tasklog.NewOutFrame(reconcileLog(resp), tasklog.TsNow))
	result := &jobResult{skipNotification: true}
	if len(resp.Failed) > 0 {
		return result, hperrors.Wrap(hperrors.ErrActionFailed).WithExtraDetail(
			"the artifacts of %d app(s) are not on every node they may start on", len(resp.Failed))
	}
	return result, nil
}

// reconcileLog is what a run did: how many apps have their artifacts, and each
// that does not, by key, with why.
func reconcileLog(resp *artifactservice.ReconcileAllResp) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "The artifacts of %d app(s) are on every node they may start on.\n", resp.Apps)
	for _, failed := range resp.Failed {
		reason := failed.Err.Error()
		if info, _ := hperrors.ParseError(failed.Err, translation.GetDefaultLang()); info != nil && info.Detail != "" {
			reason = info.Detail
		}
		fmt.Fprintf(&sb, "%s: %s\n", failed.App.Key, strings.TrimSpace(reason))
	}
	return sb.String()
}
