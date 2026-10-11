package appimagescanuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// ScanAppImages queues a scan of the app's images, scanning again those
// scanned within the hour. A scan of the app queued before and not started is
// the one answered.
func (uc *UC) ScanAppImages(
	ctx context.Context,
	_ *basedto.Auth,
	req *appimagescandto.ScanAppImagesReq,
) (*appimagescandto.ScanAppImagesResp, error) {
	app, err := uc.loadApp(ctx, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	if app.Project != nil && app.Project.Key == base.HivepaasProjectKey {
		return nil, hperrors.NewUnsupported("Scanning HivePaaS's own apps")
	}
	recorded, err := uc.imageScanService.RecordScan(ctx, uc.db, &imagescanservice.RecordScanReq{
		App: app, Trigger: base.ImageScanTriggerManual, Force: true,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !recorded.Pending {
		if err = uc.taskQueue.ScheduleTask(ctx, recorded.Task); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &appimagescandto.ScanAppImagesResp{
		Data: &appimagescandto.ScanAppImagesDataResp{Task: &basedto.ObjectIDResp{ID: recorded.Task.ID}},
	}, nil
}
