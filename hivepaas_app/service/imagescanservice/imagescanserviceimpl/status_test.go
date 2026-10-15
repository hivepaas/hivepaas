package imagescanserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

// noScanSetting is an installation where scanning was never configured.
type noScanSetting struct{ repository.SettingRepo }

func (noScanSetting) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	return nil, hperrors.Wrap(hperrors.ErrNotFound)
}

// scanHistory is the last scan done and the last tried.
type scanHistory struct {
	repository.ImageScanRepo
	done, tried *entity.ImageScan
}

func (r scanHistory) LatestScan(context.Context, database.IDB) (*entity.ImageScan, error) {
	return r.done, nil
}

func (r scanHistory) LatestAttempt(context.Context, database.IDB) (*entity.ImageScan, error) {
	return r.tried, nil
}

// Scans failing since the last that worked say why; the last that worked
// still gives the DB it used.
func TestTheStatusSaysWhyTheLastScanFailed(t *testing.T) {
	worked := &entity.ImageScan{Status: base.ImageScanStatusDone, ScannedAt: time.Unix(100, 0),
		TrivyVersion: "0.70.0"}
	failed := &entity.ImageScan{Status: base.ImageScanStatusFailed, Error: "the DB mirror is unreachable"}
	svc := &service{settingRepo: noScanSetting{}, imageScanRepo: scanHistory{done: worked, tried: failed}}

	status, err := svc.Status(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, "the DB mirror is unreachable", status.LastError)
	assert.Equal(t, "0.70.0", status.TrivyVersion)
	assert.Equal(t, time.Unix(100, 0), status.LastScanAt)
}

func TestTheStatusHasNoErrorOnceAScanWorksAgain(t *testing.T) {
	worked := &entity.ImageScan{Status: base.ImageScanStatusDone}
	svc := &service{settingRepo: noScanSetting{}, imageScanRepo: scanHistory{done: worked, tried: worked}}

	status, err := svc.Status(context.Background(), nil)

	assert.NoError(t, err)
	assert.Empty(t, status.LastError)
}

// Before any scan worked - every one failed - the error is still told.
func TestTheStatusSaysWhyWhenNoScanEverWorked(t *testing.T) {
	failed := &entity.ImageScan{Status: base.ImageScanStatusFailed, Error: "out of memory"}
	svc := &service{settingRepo: noScanSetting{}, imageScanRepo: scanHistory{tried: failed}}

	status, err := svc.Status(context.Background(), nil)

	assert.NoError(t, err)
	assert.Equal(t, "out of memory", status.LastError)
	assert.True(t, status.LastScanAt.IsZero())
}
