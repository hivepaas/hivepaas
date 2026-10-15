// Package imagescanservice scans the images apps run for known
// vulnerabilities, with Trivy: a container per scan, started on the node the
// chosen volume decides, which keeps Trivy's DBs and the layers it analyzed.
// See docs/superpowers/specs/2026-10-11-image-scanning-design.md.
package imagescanservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Service interface {
	// Validate refuses a configuration scans could not run by: switched on
	// without a volume, or with one that is gone, or memory.
	Validate(ctx context.Context, db database.IDB, cfg *entity.ImageScanSettings) error

	// Status says where scans run and what the last one found, for the settings
	// screen. It never fails the screen.
	Status(ctx context.Context, db database.IDB) (*Status, error)

	// Settings is the stored configuration and its setting; both nil when it
	// was never saved.
	Settings(ctx context.Context, db database.IDB) (*entity.Setting, *entity.ImageScanSettings, error)

	// RecordScan writes the task that scans an app's images, unscheduled: a
	// task is picked up only once the transaction it was written in has
	// committed. A scan of the app not started yet is the one asked for, and no
	// second is written. Refused while scanning is off.
	RecordScan(ctx context.Context, db database.IDB, req *RecordScanReq) (*RecordScanResp, error)

	// ScanApp scans the images the app's running tasks run, and keeps what it
	// found. The results are written on their own, not in the caller's
	// transaction: a task holds one open for as long as it runs.
	ScanApp(ctx context.Context, req *ScanAppReq) (*entity.TaskImageScanOutput, error)

	// RunDaily scans again what every app runs, removes what no app runs any
	// more, and notifies of what is new.
	RunDaily(ctx context.Context, db database.IDB, req *DailyReq) (*DailyResp, error)

	// SBOM is the packages of an image the app's tasks run, as CycloneDX JSON:
	// read as a scan reads it - on the same node, with the same volume, one at
	// a time - and not kept.
	SBOM(ctx context.Context, req *SBOMReq) ([]byte, error)
}

type SBOMReq struct {
	// App has its project and env loaded.
	App     *entity.App
	ImageID string
}

// Status is where scans run, and what the last one used.
type Status struct {
	Enabled bool
	// Volume is the chosen one; nil when none is.
	Volume *VolumeStatus
	// Image is the Trivy image scans run.
	Image string
	// DBUpdatedAt is the DB's of the last scan, zero before any.
	DBUpdatedAt  time.Time
	LastScanAt   time.Time
	TrivyVersion string
	// LastError is the last scan's, when it failed.
	LastError string
}

// VolumeStatus is the volume Trivy's data is on, and so where scans run.
type VolumeStatus struct {
	ID   string
	Name string
	// Missing says the volume is gone: no scan can run.
	Missing bool
	// NodeID and NodeName are the node a volume pinned to one is on, where
	// every scan runs; NodeLabel the label of one pinned by label.
	NodeID    string
	NodeName  string
	NodeLabel string
	// Shared says every node reaches the volume: each scan runs on the node
	// running its image.
	Shared bool
}

type RecordScanReq struct {
	App     *entity.App
	Trigger base.ImageScanTrigger
	// Force scans an image again though it was scanned within the hour.
	Force bool
}

type RecordScanResp struct {
	Task *entity.Task
	// Pending says Task was already waiting to start: it is scheduled already.
	Pending bool
}

type ScanAppReq struct {
	// App has its project and env loaded.
	App     *entity.App
	Trigger base.ImageScanTrigger
	Force   bool
	// LogStore gets what the scan says, Trivy's own lines among it.
	LogStore *tasklog.Store
}

type DailyReq struct {
	*queue.TaskExecData
	Setting *entity.Setting
}

type DailyResp struct {
	// SkipResultNotification says the job's own notification is not sent: a
	// run that succeeded notifies only of what is new, with its own message.
	SkipResultNotification bool
}
