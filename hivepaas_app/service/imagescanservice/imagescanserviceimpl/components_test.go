package imagescanserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// bare is a cluster that runs none of HivePaaS's parts: no service of its
// stack, no system app, no node.
type bare struct{ docker.Manager }

func (bare) ServiceInspect(context.Context, string, ...docker.ServiceInspectOption) (
	*client.ServiceInspectResult, error) {
	return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
}

func (bare) NodeList(context.Context, ...docker.NodeListOption) (*client.NodeListResult, error) {
	return &client.NodeListResult{}, nil
}

type noSystemApps struct{ systemappservice.Service }

func (noSystemApps) LoadApp(context.Context, database.IDB, string) (*entity.App, error) {
	return nil, nil
}

// componentScans records what is kept of the parts' results.
type componentScans struct {
	repository.ImageScanRepo
	kept []string
}

func (r *componentScans) DeleteComponentScansExcept(_ context.Context, _ database.IDB, keep []string) error {
	r.kept = keep
	return nil
}

// A part that does not run is not scanned, and what it ran before goes.
func TestScanComponentsPassesOverWhatDoesNotRun(t *testing.T) {
	scans := &componentScans{}
	svc := &service{dockerManager: bare{}, systemAppService: noSystemApps{}, imageScanRepo: scans}
	scanned, failed, err := svc.scanComponents(context.Background(), nil,
		&imageScanReq{log: tasklog.NewNullStore()})
	assert.NoError(t, err)
	assert.Zero(t, scanned)
	assert.Zero(t, failed)
	assert.Equal(t, []string{}, scans.kept, "every part's results go")
}
