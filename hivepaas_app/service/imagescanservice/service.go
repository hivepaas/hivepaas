// Package imagescanservice scans the images apps run for known
// vulnerabilities, with Trivy: a server in the hidden hivepaas project holds
// the vulnerability DB and the layers it has analyzed; a runner on every node
// reads that node's images and asks the server. See
// docs/superpowers/specs/2026-10-11-image-scanning-design.md.
package imagescanservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Service interface {
	// Apply makes the cluster match the stored configuration: it provisions the
	// server and the runner on the save that switches scanning on, reconciles
	// them on the rest, and removes both on the one that switches it off - the
	// results are kept. It is idempotent, so a failed save leaves work the next
	// one retries.
	Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)

	// Sync brings the apps to the stored settings, as a save does, and further:
	// an app whose service is gone is removed and provisioned again. It runs in
	// the caller's transaction.
	Sync(ctx context.Context, db database.IDB) (*SyncResp, error)

	// Status says what runs, for the settings screen. It never fails the screen.
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
}

type SettingApplyReq struct {
	// Setting is the row just saved. Apply loads it when nil.
	Setting       *entity.Setting
	TriggerUserID string
	// ScanResources is what the runner - every scan - runs under. Nil leaves a
	// running one as it is, and gives a new one the defaults.
	ScanResources *systemappservice.Resources
}

type SettingApplyResp struct {
	// Tasks are the deployments the apply queued, for the caller to schedule
	// once its transaction committed.
	Tasks []*entity.Task
	// Cleanup removes from docker what provisioning created there, for the
	// caller to run when its transaction does not commit.
	Cleanup func(ctx context.Context) error
	// Removed says the apps were taken down by this call.
	Removed bool
}

// SyncResp is what Sync found and did.
type SyncResp struct {
	Apps    []*entity.SystemAppSyncOutput
	Tasks   []*entity.Task
	Cleanup func(ctx context.Context) error
}

// Status is what runs.
type Status struct {
	Enabled bool
	// ServerAppID and RunnerAppID are set when the apps exist.
	ServerAppID string
	RunnerAppID string
	// ServerRunning says the server has a task running.
	ServerRunning bool
	// RunnerNodes is on how many nodes the runner runs, of Nodes.
	RunnerNodes int
	Nodes       int
	// Image is the Trivy image the apps run.
	Image string
	// DBUpdatedAt is the DB's of the last scan, zero before any.
	DBUpdatedAt  time.Time
	LastScanAt   time.Time
	TrivyVersion string
	// ScanResources is what the runner runs under, or what a new one would get.
	ScanResources systemappservice.Resources
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
