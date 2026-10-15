package scannedapps

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appimagescanuc/appimagescandto"
)

// FillDetail writes what an app - or a part of HivePaaS - ran when last
// scanned into data: its images, their findings no ignore sets aside, the worst
// first, and those one does, each with its ignore.
func FillDetail(
	ctx context.Context,
	db database.IDB,
	imageScanRepo repository.ImageScanRepo,
	appScan *entity.AppImageScan,
	ignores []*imagescanservice.Ignore,
	data *appimagescandto.AppImageScanResp,
) error {
	data.ScannedAt, data.Error = &appScan.ScannedAt, appScan.Error
	scans, err := imageScanRepo.ListScans(ctx, db, appScan.ImageIDs, true)
	if err != nil {
		return hperrors.Wrap(err)
	}

	now := timeutil.NowUTC()
	var total imagescanservice.Summary
	var findings []*entity.ImageScanFinding
	for _, imageID := range appScan.ImageIDs {
		for _, scan := range scans {
			if scan.ImageID != imageID {
				continue
			}
			summary := imagescanservice.Summarize(scan.Findings, ignores, now)
			total.Merge(summary)
			data.Images = append(data.Images, appimagescandto.TransformImage(scan, summary))
			findings = append(findings, scan.Findings...)
		}
	}
	appimagescandto.SortFindings(findings)
	for _, f := range findings {
		if ignore := imagescanservice.Applied(f, ignores, now); ignore != nil {
			data.Ignored = append(data.Ignored, &appimagescandto.IgnoredFindingResp{
				FindingResp: appimagescandto.TransformFinding(f),
				Ignore:      appimagescandto.TransformIgnoreRef(ignore),
			})
			continue
		}
		data.Findings = append(data.Findings, appimagescandto.TransformFinding(f))
	}
	data.Summary = appimagescandto.TransformSummary(total)
	return nil
}
