package sysbackupserviceimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysbackupservice"
	"github.com/hivepaas/hivepaas/services/backup"
)

const (
	// systemBackupSource is what every system backup is recorded under: the
	// repository's retention counts it apart from everything else.
	systemBackupSource = "hivepaas@system-backup:/system"
	systemBackupTag    = entity.DataBackupTagSource + ":system-backup"
	// The files of a system backup's snapshot.
	dbDumpFileName = "db.pg_dump"
	specFileName   = "spec.tar.gz"
)

func (s *service) Backup(
	ctx context.Context,
	db database.Tx,
	req *sysbackupservice.SysBackupReq,
) (resp *sysbackupservice.SysBackupResp, err error) {
	defer safego.RecoverTo(&err)

	start := time.Now()
	settings := req.SysBackupSettings
	includes := settings.Includes()
	if len(includes) == 0 {
		return nil, hperrors.NewArgumentInvalid("includeDB").WithExtraDetail("the system backup takes nothing")
	}
	target, err := s.target(ctx, db, settings)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	dir, err := s.workDir()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer os.RemoveAll(dir)
	snapshotDir := filepath.Join(dir, "snapshot")
	if err = os.Mkdir(snapshotDir, base.DirModeDefault); err != nil {
		return nil, hperrors.Wrap(err)
	}

	if settings.IncludeDB {
		logTo(ctx, req, "Dumping the database")
		if err = s.dumpDB(ctx, filepath.Join(snapshotDir, dbDumpFileName), req.LogStore); err != nil {
			return nil, hperrors.Wrap(fmt.Errorf("dumping the database: %w", err))
		}
	}
	if settings.IncludeSpec {
		logTo(ctx, req, "Exporting the spec")
		if err = s.exportSpec(ctx, db, settings, dir, snapshotDir); err != nil {
			return nil, hperrors.Wrap(fmt.Errorf("exporting the spec: %w", err))
		}
	}

	logTo(ctx, req, "Taking the snapshot")
	taken, err := s.backupRepoService.BackupLocalDirectory(ctx, db, &backupreposervice.BackupLocalDirectoryReq{
		RepoTarget: *target, Dir: snapshotDir, Source: systemBackupSource,
		Description: fmt.Sprintf("System backup (run %s)", req.Task.ID),
		Tags:        []string{systemBackupTag, entity.DataBackupTagRun + ":" + req.Task.ID},
		Progress:    func(msg string) { logTo(ctx, req, "%s", msg) },
	})
	if err != nil {
		s.deleteRunSnapshots(context.WithoutCancel(ctx), db, req, target)
		return nil, hperrors.Wrap(err)
	}
	if taken == nil || taken.Snapshot == nil || taken.Snapshot.Snapshot == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("the engine named no snapshot")
	}

	output := &entity.TaskSystemBackupOutput{
		SnapshotID: taken.Snapshot.Snapshot.ID, SizeBytes: taken.Snapshot.Snapshot.SizeBytes, Includes: includes,
	}
	logTo(ctx, req, "Snapshot %s taken: %s in %s", output.SnapshotID, unit.DataSize(output.SizeBytes).String(),
		time.Since(start).Truncate(time.Second))
	if err = s.syncSnapshots(ctx, db, target); err != nil {
		// The snapshot is taken: its record comes with the repository's next sync.
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"The repository's snapshot list could not be updated: "+err.Error()+"\n", tasklog.TsNow))
	}
	req.Task.MustSetOutput(output)
	return &sysbackupservice.SysBackupResp{}, nil
}

// target is the repository the backup goes into, which must be active.
func (s *service) target(
	ctx context.Context,
	db database.IDB,
	settings *entity.SystemBackup,
) (*backupreposervice.RepoTarget, error) {
	repo, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeBackupRepo, settings.TargetRepository.ID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	scope, err := s.repoScope(ctx, db, repo)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &backupreposervice.RepoTarget{Scope: scope, RepoSetting: repo}, nil
}

// exportSpec exports the whole installation's spec into snapshotDir, as the
// export names it: spec.tar.gz, and .age after it when it is encrypted.
func (s *service) exportSpec(
	ctx context.Context,
	db database.IDB,
	settings *entity.SystemBackup,
	dir, snapshotDir string,
) error {
	workDir := filepath.Join(dir, "spec-work")
	if err := os.Mkdir(workDir, base.DirModeDefault); err != nil {
		return hperrors.Wrap(err)
	}
	mode := specmodel.SecretsMode(settings.SpecSecrets)
	passphrase := ""
	if mode == specmodel.SecretsModeEncrypted {
		var err error
		if passphrase, err = settings.SpecPassphrase.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	exported, err := s.specService.Export(ctx, db, &specservice.ExportReq{
		Scope: &entity.ObjectScope{ScopeType: base.ObjectScopeGlobal}, SecretsMode: mode, Passphrase: passphrase,
		WorkDir: workDir,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	name := specFileName
	if filepath.Ext(exported.Filename) == ".age" {
		name += ".age"
	}
	return hperrors.Wrap(os.Rename(exported.Path, filepath.Join(snapshotDir, name)))
}

// deleteRunSnapshots deletes what a failed run left in the repository: the
// snapshots tagged with the run.
func (s *service) deleteRunSnapshots(
	ctx context.Context,
	db database.IDB,
	req *sysbackupservice.SysBackupReq,
	target *backupreposervice.RepoTarget,
) {
	repo, err := target.RepoSetting.AsBackupRepo()
	if err != nil {
		return
	}
	listed, err := s.backupRepoService.ListSnapshots(ctx, db, &backupreposervice.ListSnapshotsReq{
		Scope: target.Scope, Repo: repo, RepoID: target.RepoSetting.ID,
		Options: &backup.ListSnapshotsOptions{Tags: []string{entity.DataBackupTagRun + ":" + req.Task.ID}},
	})
	if err != nil {
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Could not look for a partial snapshot to delete: "+err.Error()+"\n", tasklog.TsNow))
		return
	}
	for _, item := range listed.Snapshots {
		if item.Snapshot == nil {
			continue
		}
		err = s.backupRepoService.DeleteSnapshot(ctx, db, &backupreposervice.DeleteSnapshotReq{
			RepoTarget: *target, SnapshotID: item.Snapshot.ID,
		})
		if err != nil {
			_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Could not delete the partial snapshot %s: %v\n", item.Snapshot.ID, err), tasklog.TsNow))
			continue
		}
		logTo(ctx, req, "Deleted the partial snapshot %s", item.Snapshot.ID)
	}
}

// syncSnapshots makes the repository's snapshot list match what it now holds.
func (s *service) syncSnapshots(ctx context.Context, db database.Tx, target *backupreposervice.RepoTarget) error {
	repo, err := target.RepoSetting.AsBackupRepo()
	if err != nil {
		return hperrors.Wrap(err)
	}
	listed, err := s.backupRepoService.ListSnapshots(ctx, db, &backupreposervice.ListSnapshotsReq{
		Scope: target.Scope, Repo: repo, RepoID: target.RepoSetting.ID,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = s.backupRepoService.SyncRepoSnapshots(ctx, db, &backupreposervice.SyncRepoSnapshotsReq{
		Scope: target.Scope, RepoSetting: target.RepoSetting, Remaining: listed.Snapshots,
	})
	return hperrors.Wrap(err)
}

func logTo(ctx context.Context, req *sysbackupservice.SysBackupReq, format string, args ...any) {
	_ = req.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf(format, args...)+"\n", tasklog.TsNow))
}
