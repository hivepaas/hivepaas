package imagescanserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// scanningOn is scanning switched on with a volume, and the volumes it holds.
type scanningOn struct {
	settingRepo
	volumeID string
}

func (r *scanningOn) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	setting := &entity.Setting{ID: "scan", Type: base.SettingTypeImageScan}
	setting.MustSetData(&entity.ImageScanSettings{Enabled: true, Volume: entity.ObjectID{ID: r.volumeID}})
	return setting, nil
}

// labeledNodes answers the ready nodes carrying a label: none.
type labeledNodes struct {
	agentservice.Service
	nodes []string
}

func (a labeledNodes) NodeIDsWithLabel(context.Context, string) ([]string, error) {
	return a.nodes, nil
}

// An app's scan that cannot run - its volume is gone - leaves the app saying
// why, with what it ran last, rather than a failed task alone.
func TestAScanWithNoVolumeToRunOnSaysWhyOnTheApp(t *testing.T) {
	repo := &appScans{previous: &entity.AppImageScan{AppID: "app", ImageIDs: []string{imageA}}}
	svc := &service{settingRepo: &scanningOn{settingRepo: settingRepo{}, volumeID: "gone"},
		imageScanRepo: repo}

	out, err := svc.ScanApp(context.Background(), &imagescanservice.ScanAppReq{App: &entity.App{ID: "app"}})

	assert.ErrorIs(t, err, hperrors.ErrImageScanNotReady)
	if assert.NotNil(t, repo.written) {
		assert.Contains(t, repo.written.Error, "no longer exists")
		assert.Equal(t, []string{imageA}, repo.written.ImageIDs, "what it ran last is kept")
		assert.Equal(t, repo.written.Error, out.Error)
	}
}

// A volume pinned by a label no ready node carries is said on the status:
// no scan can run until one does.
func TestTheStatusSaysNoNodeCarriesTheVolumesLabel(t *testing.T) {
	volumes := map[string]*entity.ClusterVolume{"lab": {NodeLabel: "scan=yes", Managed: true, Driver: "local"}}
	svc := &service{settingRepo: &scanningOn{settingRepo: settingRepo{volumes: volumes}, volumeID: "lab"},
		imageScanRepo: scanHistory{}, agentService: labeledNodes{}}

	status, err := svc.Status(context.Background(), nil)
	if assert.NoError(t, err) && assert.NotNil(t, status.Volume) {
		assert.Equal(t, `No ready node carries the label "scan=yes": no scan can run.`, status.Volume.Problem)
	}

	svc.agentService = labeledNodes{nodes: []string{"n1"}}
	status, err = svc.Status(context.Background(), nil)
	if assert.NoError(t, err) && assert.NotNil(t, status.Volume) {
		assert.Empty(t, status.Volume.Problem, "a node carries it")
	}
}
