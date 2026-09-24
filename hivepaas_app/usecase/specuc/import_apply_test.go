package specuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

type fakeApplyService struct {
	fakeSpecService
	resp *specservice.ApplyImportResp
	err  error
}

func (f *fakeApplyService) ApplyImport(
	_ context.Context, _ database.IDB, _ *specservice.ApplyImportReq,
) (*specservice.ApplyImportResp, error) {
	return f.resp, f.err
}

type fakeQueue struct {
	queue.TaskQueue
	scheduled []*entity.Task
}

func (f *fakeQueue) ScheduleTask(_ context.Context, tasks ...*entity.Task) error {
	f.scheduled = append(f.scheduled, tasks...)
	return nil
}

func applyResp() *specservice.ApplyImportResp {
	return &specservice.ApplyImportResp{
		Plan: &specmodel.ImportPlan{
			Bundle:  specmodel.BundleInfo{Digest: "d1gest", Scope: "global", SecretsMode: specmodel.SecretsModeOmit},
			Summary: map[string]int{"create": 2},
		},
		AfterCommit: func(context.Context, database.IDB) error { return nil },
		Tasks:       []*entity.Task{{ID: "task_1"}},
		Deployments: []*specservice.ImportDeployment{{AppID: "a1", DeploymentID: "d1"}},
	}
}

func applyImportReq() *specdto.ApplyImportReq {
	req := specdto.NewApplyImportReq()
	req.Scope = entity.NewObjectScopeGlobal()
	req.Bundle, req.PlanHash, req.AcceptIssues = []byte("bundle"), "h4sh", true
	req.Selection.Include = []string{"projects/a"}
	_ = req.ModifyRequest()
	return req
}

// An import is recorded with the bundle it came from, what it chose and what
// it did.
func TestApplyImportIsRecorded(t *testing.T) {
	uc, audit, _ := newTestUC(t)
	uc.specService = &fakeApplyService{resp: applyResp()}
	req := applyImportReq()

	_, err := uc.applyInTx(context.Background(), nil, adminAuth(), req, &specservice.ApplyImportReq{})

	assert.NoError(t, err)
	entries := entriesOfType(audit, base.AuditLogTypeSpecImport)
	if assert.Len(t, entries, 1) {
		for _, want := range []string{"d1gest", "projects/a", `"create":2`, "update"} {
			assert.Contains(t, entries[0].Detail, want)
		}
	}
}

// What provisioning made in docker comes back with the error, for the
// transaction's caller to remove.
func TestApplyImportReturnsWhatToCleanUpWithTheError(t *testing.T) {
	uc, audit, _ := newTestUC(t)
	resp := applyResp()
	uc.specService = &fakeApplyService{resp: resp, err: errors.New("provisioning failed")}

	applied, err := uc.applyInTx(context.Background(), nil, adminAuth(), applyImportReq(), &specservice.ApplyImportReq{})

	assert.Error(t, err)
	assert.Same(t, resp, applied)
	assert.Empty(t, entriesOfType(audit, base.AuditLogTypeSpecImport), "nothing is recorded")
}

// Once committed, the deployments are scheduled after phase 2, and a failure
// there is a warning: the import stands.
func TestApplyImportSchedulesAfterPhaseTwo(t *testing.T) {
	uc, _, _ := newTestUC(t)
	tasks := &fakeQueue{}
	uc.taskQueue = tasks
	resp := applyResp()
	var order []string
	resp.AfterCommit = func(context.Context, database.IDB) error {
		order = append(order, "phase 2")
		assert.Empty(t, tasks.scheduled, "nothing is scheduled before phase 2")
		return errors.New("disk full")
	}

	out := uc.afterImport(context.Background(), nil, resp)

	assert.Equal(t, []string{"phase 2"}, order)
	assert.Equal(t, resp.Tasks, tasks.scheduled)
	assert.Contains(t, out.Meta.Warning, "disk full")
	assert.Equal(t, resp.Deployments, out.Data.Deployments)
}
