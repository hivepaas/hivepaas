package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// imageScanBatch caps how many findings one insert carries: Postgres takes at
// most 65535 parameters a statement, and a finding is eleven.
const imageScanBatch = 1000

// ImageScanRepo keeps the scans of images and what apps ran. A scan is
// replaced whole: its row, and its findings with it.
type ImageScanRepo interface {
	// ListScans are the scans of the images, findings loaded when withFindings.
	ListScans(ctx context.Context, db database.IDB, imageIDs []string, withFindings bool,
		opts ...bunex.SelectQueryOption) ([]*entity.ImageScan, error)
	// ListFindings are the findings of the images.
	ListFindings(ctx context.Context, db database.IDB, imageIDs []string,
		opts ...bunex.SelectQueryOption) ([]*entity.ImageScanFinding, error)
	// LatestScan is the scan done last; nil before any.
	LatestScan(ctx context.Context, db database.IDB) (*entity.ImageScan, error)
	// LatestAttempt is the scan tried last, done or failed; nil before any.
	LatestAttempt(ctx context.Context, db database.IDB) (*entity.ImageScan, error)
	// ReplaceScan writes a scan and its findings in place of the image's last.
	ReplaceScan(ctx context.Context, db database.IDB, scan *entity.ImageScan,
		findings []*entity.ImageScanFinding) error
	// MarkFailed says an image's last scan failed, keeping what the one before
	// found.
	MarkFailed(ctx context.Context, db database.IDB, scan *entity.ImageScan) error

	GetAppScan(ctx context.Context, db database.IDB, appID string) (*entity.AppImageScan, error)
	ListAppScans(ctx context.Context, db database.IDB, opts ...bunex.SelectQueryOption) (
		[]*entity.AppImageScan, error)
	UpsertAppScan(ctx context.Context, db database.IDB, appScan *entity.AppImageScan) error

	// DeleteGone removes what no app is there for any more: the rows of apps
	// that were deleted, and the scans of images no app has run since before.
	// HivePaaS's components are no apps: their rows are left alone.
	DeleteGone(ctx context.Context, db database.IDB, before time.Time) (apps, images int, err error)
	// DeleteComponentScansExcept removes the rows of HivePaaS's components but
	// those of keep: the parts that run.
	DeleteComponentScansExcept(ctx context.Context, db database.IDB, keep []string) error
}

type imageScanRepo struct {
}

func NewImageScanRepo() ImageScanRepo {
	return &imageScanRepo{}
}

func (repo *imageScanRepo) ListScans(ctx context.Context, db database.IDB, imageIDs []string,
	withFindings bool, opts ...bunex.SelectQueryOption) ([]*entity.ImageScan, error) {
	if len(imageIDs) == 0 {
		return nil, nil
	}
	var scans []*entity.ImageScan
	query := db.NewSelect().Model(&scans).Where("image_scan.image_id IN (?)", bun.List(imageIDs))
	if withFindings {
		query = query.Relation("Findings")
	}
	query = bunex.ApplySelect(query, opts...)
	if err := query.Scan(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return scans, nil
}

func (repo *imageScanRepo) ListFindings(ctx context.Context, db database.IDB, imageIDs []string,
	opts ...bunex.SelectQueryOption) ([]*entity.ImageScanFinding, error) {
	if len(imageIDs) == 0 {
		return nil, nil
	}
	var findings []*entity.ImageScanFinding
	query := db.NewSelect().Model(&findings).
		Where("image_scan_finding.image_id IN (?)", bun.List(imageIDs))
	query = bunex.ApplySelect(query, opts...)
	if err := query.Scan(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return findings, nil
}

func (repo *imageScanRepo) LatestScan(ctx context.Context, db database.IDB) (*entity.ImageScan, error) {
	var scans []*entity.ImageScan
	err := db.NewSelect().Model(&scans).
		Where("image_scan.status = ?", base.ImageScanStatusDone).
		Order("image_scan.scanned_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(scans) == 0 {
		return nil, nil
	}
	return scans[0], nil
}

func (repo *imageScanRepo) LatestAttempt(ctx context.Context, db database.IDB) (*entity.ImageScan, error) {
	var scans []*entity.ImageScan
	// A failed scan writes its updated_at and keeps the scanned_at of the one
	// before it: updated_at is when each was tried.
	err := db.NewSelect().Model(&scans).
		Order("image_scan.updated_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(scans) == 0 {
		return nil, nil
	}
	return scans[0], nil
}

func (repo *imageScanRepo) MarkFailed(ctx context.Context, db database.IDB, scan *entity.ImageScan) error {
	_, err := db.NewInsert().Model(scan).
		On("CONFLICT (image_id) DO UPDATE").
		Set("status = EXCLUDED.status").
		Set("error = EXCLUDED.error").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *imageScanRepo) ReplaceScan(ctx context.Context, db database.IDB, scan *entity.ImageScan,
	findings []*entity.ImageScanFinding) error {
	_, err := db.NewInsert().Model(scan).
		On("CONFLICT (image_id) DO UPDATE").
		Set("image_ref = EXCLUDED.image_ref").
		Set("platform = EXCLUDED.platform").
		Set("os_family = EXCLUDED.os_family").
		Set("os_name = EXCLUDED.os_name").
		Set("os_eosl = EXCLUDED.os_eosl").
		Set("status = EXCLUDED.status").
		Set("source = EXCLUDED.source").
		Set("error = EXCLUDED.error").
		Set("trivy_version = EXCLUDED.trivy_version").
		Set("db_updated_at = EXCLUDED.db_updated_at").
		Set("scanned_at = EXCLUDED.scanned_at").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = db.NewDelete().Model((*entity.ImageScanFinding)(nil)).
		Where("image_id = ?", scan.ImageID).Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for start := 0; start < len(findings); start += imageScanBatch {
		batch := findings[start:min(start+imageScanBatch, len(findings))]
		// Trivy reports a vulnerability once per place a package is found; two
		// findings of one key are one row.
		if _, err = db.NewInsert().Model(&batch).On("CONFLICT DO NOTHING").Exec(ctx); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (repo *imageScanRepo) GetAppScan(ctx context.Context, db database.IDB, appID string) (
	*entity.AppImageScan, error) {
	appScans, err := repo.ListAppScans(ctx, db, bunex.SelectWhere("app_image_scan.app_id = ?", appID))
	if err != nil {
		return nil, err
	}
	if len(appScans) == 0 {
		return nil, hperrors.NewNotFound("AppImageScan")
	}
	return appScans[0], nil
}

func (repo *imageScanRepo) ListAppScans(ctx context.Context, db database.IDB,
	opts ...bunex.SelectQueryOption) ([]*entity.AppImageScan, error) {
	var appScans []*entity.AppImageScan
	query := db.NewSelect().Model(&appScans)
	query = bunex.ApplySelect(query, opts...)
	if err := query.Scan(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return appScans, nil
}

func (repo *imageScanRepo) UpsertAppScan(ctx context.Context, db database.IDB,
	appScan *entity.AppImageScan) error {
	_, err := db.NewInsert().Model(appScan).
		On("CONFLICT (app_id) DO UPDATE").
		Set("image_ids = EXCLUDED.image_ids").
		Set("error = EXCLUDED.error").
		Set("scanned_at = EXCLUDED.scanned_at").
		Set("reported = EXCLUDED.reported").
		Exec(ctx)
	return hperrors.Wrap(err)
}

func (repo *imageScanRepo) DeleteGone(ctx context.Context, db database.IDB, before time.Time) (
	apps, images int, err error) {
	res, err := db.NewDelete().Model((*entity.AppImageScan)(nil)).
		Where("NOT EXISTS (SELECT 1 FROM apps WHERE apps.id = app_image_scan.app_id AND apps.deleted_at IS NULL)").
		Where("app_image_scan.app_id NOT LIKE ?", componentIDPrefix+"%").
		Exec(ctx)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	deletedApps, _ := res.RowsAffected()

	res, err = db.NewDelete().Model((*entity.ImageScan)(nil)).
		Where("image_scan.scanned_at < ?", before).
		Where("NOT EXISTS (SELECT 1 FROM app_image_scans AS a WHERE image_scan.image_id = ANY(a.image_ids))").
		Exec(ctx)
	if err != nil {
		return 0, 0, hperrors.Wrap(err)
	}
	deletedImages, _ := res.RowsAffected()
	return int(deletedApps), int(deletedImages), nil
}

// componentIDPrefix starts the IDs HivePaaS's components are kept under:
// imagescanservice.ComponentIDPrefix, which this package cannot import.
const componentIDPrefix = "hivepaas:"

func (repo *imageScanRepo) DeleteComponentScansExcept(ctx context.Context, db database.IDB, keep []string) error {
	query := db.NewDelete().Model((*entity.AppImageScan)(nil)).
		Where("app_image_scan.app_id LIKE ?", componentIDPrefix+"%")
	if len(keep) > 0 {
		query = query.Where("app_image_scan.app_id NOT IN (?)", bun.List(keep))
	}
	_, err := query.Exec(ctx)
	return hperrors.Wrap(err)
}
