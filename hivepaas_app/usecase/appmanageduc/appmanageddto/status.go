package appmanageddto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

type GetManagedStatusReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetManagedStatusReq() *GetManagedStatusReq {
	return &GetManagedStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetManagedStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetManagedStatusResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *ManagedStatusResp `json:"data"`
}

// ManagedStatusResp is a managed database's state, read from its jobs' last
// runs.
type ManagedStatusResp struct {
	Engine    string                `json:"engine"`
	EngineAPI int                   `json:"engineApi"`
	Archiving *ManagedArchivingResp `json:"archiving"`
	// RestoreWindow is what a restore may target; null before the first base
	// backup.
	RestoreWindow *ManagedTimeWindowResp `json:"restoreWindow"`
	LastBackup    *ManagedBackupResp     `json:"lastBackup"`
	// Backups are the base backups kept, oldest first: what a restore to a backup
	// may name.
	Backups         []*ManagedBackupRefResp `json:"backups"`
	LastRestoreTest *ManagedRestoreTestResp `json:"lastRestoreTest"`
	// LastDump is the dump's last run, null while dumps are off or before the
	// first.
	LastDump *ManagedDumpResp `json:"lastDump"`
	// Members are the replicas, each as the database's last health check saw it.
	Members []*ManagedMemberStatusResp `json:"members"`
	// Placement is where the archive is from the database: off-site, shared or
	// same-node, which the dashboard warns about.
	Placement string `json:"placement"`
	// Suggestions are changes worth making: verifier, to run the restore test
	// in a verifier app; dump-apart, to dump at another time than the base
	// backup, as both read the whole database.
	Suggestions []string `json:"suggestions"`
	// RestorePoints are the points recorded, oldest first.
	RestorePoints []*ManagedRestorePointResp `json:"restorePoints"`
	// HeldBack is what the kept points make retention keep beyond its own: the
	// base backups and their size, the log after them not counted. Null when
	// they keep nothing more.
	HeldBack *ManagedHeldBackResp `json:"heldBack"`
}

type ManagedHeldBackResp struct {
	Backups int   `json:"backups"`
	Bytes   int64 `json:"bytes"`
}

type ManagedArchivingResp struct {
	// State is ok, failing, or unknown before the first health check.
	State          string     `json:"state"`
	CheckedAt      *time.Time `json:"checkedAt"`
	LastArchivedAt *time.Time `json:"lastArchivedAt"`
	FailedCount    int        `json:"failedCount"`
	// PendingBytes is what waits on the node to be archived: pg_wal for
	// Postgres, the binary logs not yet pushed for MariaDB.
	PendingBytes int64 `json:"pendingBytes"`
	// Stale says the health check has not run for a while.
	Stale bool `json:"stale"`
}

type ManagedTimeWindowResp struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ManagedBackupResp struct {
	Succeeded bool       `json:"succeeded"`
	EndedAt   time.Time  `json:"endedAt"`
	Name      string     `json:"name"`
	At        *time.Time `json:"at"`
	Bytes     int64      `json:"bytes"`
}

type ManagedBackupRefResp struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
}

// ManagedDumpResp is a dump's run, and the snapshot it took when it succeeded:
// the app's backup snapshots list it.
type ManagedDumpResp struct {
	Succeeded  bool      `json:"succeeded"`
	EndedAt    time.Time `json:"endedAt"`
	SnapshotID string    `json:"snapshotId"`
	Bytes      int64     `json:"bytes"`
}

// ManagedMemberStatusResp is a replica: streaming, catchup and the like as
// pg_stat_replication says, disconnected when the database has no connection
// from it, unknown before a health check has looked; and how far behind it is,
// in bytes of WAL and in seconds, null while nothing is written.
type ManagedMemberStatusResp struct {
	App        *basedto.ObjectIDResp `json:"app"`
	Key        string                `json:"key"`
	Role       string                `json:"role"`
	Node       string                `json:"node"`
	State      string                `json:"state"`
	LagBytes   int64                 `json:"lagBytes"`
	LagSeconds *float64              `json:"lagSeconds"`
	CheckedAt  *time.Time            `json:"checkedAt"`
}

type ManagedRestoreTestResp struct {
	Passed          bool      `json:"passed"`
	EndedAt         time.Time `json:"endedAt"`
	DurationSeconds int64     `json:"durationSeconds"`
	RestoredTo      string    `json:"restoredTo"`
}

func TransformManagedStatus(engine string, engineAPI int, status *managedmodel.Status) *ManagedStatusResp {
	resp := &ManagedStatusResp{
		Engine:    engine,
		EngineAPI: engineAPI,
		Archiving: &ManagedArchivingResp{
			State:          status.Archiving.State,
			CheckedAt:      status.Archiving.CheckedAt,
			LastArchivedAt: status.Archiving.LastArchivedAt,
			FailedCount:    status.Archiving.FailedCount,
			PendingBytes:   status.Archiving.PendingBytes,
			Stale:          status.Archiving.Stale,
		},
		Placement:   status.Placement,
		Suggestions: append([]string{}, status.Suggestions...),
		Backups:     make([]*ManagedBackupRefResp, 0, len(status.Backups)),
	}
	if window := status.RestoreWindow; window != nil {
		resp.RestoreWindow = &ManagedTimeWindowResp{From: window.From, To: window.To}
	}
	if backup := status.LastBackup; backup != nil {
		resp.LastBackup = &ManagedBackupResp{Succeeded: backup.Succeeded, EndedAt: backup.EndedAt,
			Name: backup.Name, Bytes: backup.Bytes}
		if !backup.At.IsZero() {
			resp.LastBackup.At = &backup.At
		}
	}
	resp.RestorePoints = make([]*ManagedRestorePointResp, 0, len(status.RestorePoints))
	for _, point := range status.RestorePoints {
		resp.RestorePoints = append(resp.RestorePoints, TransformRestorePointStatus(point))
	}
	if held := status.HeldBack; held != nil {
		resp.HeldBack = &ManagedHeldBackResp{Backups: held.Backups, Bytes: held.Bytes}
	}
	for _, backup := range status.Backups {
		resp.Backups = append(resp.Backups, &ManagedBackupRefResp{Name: backup.Name, At: backup.At})
	}
	if test := status.LastRestoreTest; test != nil {
		resp.LastRestoreTest = &ManagedRestoreTestResp{Passed: test.Passed, EndedAt: test.EndedAt,
			DurationSeconds: int64(test.Duration / time.Second), RestoredTo: test.RestoredTo}
	}
	resp.Members = make([]*ManagedMemberStatusResp, 0, len(status.Members))
	for _, member := range status.Members {
		resp.Members = append(resp.Members, &ManagedMemberStatusResp{
			App: &basedto.ObjectIDResp{ID: member.Member.App.ID}, Key: member.Member.Key, Role: member.Member.Role,
			Node: member.Member.Node, State: member.State, LagBytes: member.LagBytes, LagSeconds: member.LagSeconds,
			CheckedAt: member.CheckedAt,
		})
	}
	if dump := status.LastDump; dump != nil {
		resp.LastDump = &ManagedDumpResp{Succeeded: dump.Succeeded, EndedAt: dump.EndedAt,
			SnapshotID: dump.SnapshotID, Bytes: dump.Bytes}
	}
	return resp
}
