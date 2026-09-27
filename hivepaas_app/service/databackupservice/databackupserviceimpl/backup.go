package databackupserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/databackupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/services/backup"
)

// dataBackupTagRun is the tag of every snapshot one run takes: what finds the
// snapshot a failed command left behind when kopia did not name it.
const dataBackupTagRun = "hivepaas.run"

// dataBackupSource and the job's ID are the source every snapshot of the job is
// recorded under, whichever machine took it: kopia's retention goes by source,
// and the backend's or an agent's hostname changes as it is redeployed.
const dataBackupSource = "hivepaas@data-backup:/"

func (s *service) Backup(
	ctx context.Context,
	db database.Tx,
	req *databackupservice.BackupReq,
) (*databackupservice.BackupResp, error) {
	start := time.Now()
	job := req.JobSetting.MustAsSchedJob()
	dataBackup := job.DataBackup
	if dataBackup == nil {
		return nil, hperrors.Wrap(hperrors.ErrSettingMissing).WithParam("Name", "data backup")
	}
	if req.App == nil {
		return nil, hperrors.NewNotFound("App")
	}
	repoSetting := req.RefObjects.RefSettings[dataBackup.TargetRepository.ID]
	if repoSetting == nil || !repoSetting.IsActive() {
		return nil, hperrors.NewNotFound("Active backup repository")
	}
	scope, err := s.repoScope(ctx, db, repoSetting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	target := backupreposervice.RepoTarget{Scope: scope, RepoSetting: repoSetting, RefObjects: req.RefObjects}
	tags := append(dataBackup.SnapshotTags(req.JobSetting.ID, req.App.ID), dataBackupTagRun+":"+req.Task.ID)
	description := fmt.Sprintf("%s (run %s)", req.JobSetting.Name, req.Task.ID)

	var snapshot *backupreposervice.RepoSnapshot
	switch dataBackup.Source {
	case base.SchedJobDataBackupSourceCommand:
		snapshot, err = s.backupCommand(ctx, db, req, dataBackup, target, tags, description)
	case base.SchedJobDataBackupSourceVolume:
		snapshot, err = s.backupVolume(ctx, db, req, dataBackup, target, tags, description)
	default:
		err = hperrors.NewUnsupported(fmt.Sprintf("Data backup source '%s'", dataBackup.Source))
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if snapshot == nil || snapshot.Snapshot == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("the engine named no snapshot")
	}

	result := &entity.SchedJobDataBackupResult{SnapshotID: snapshot.Snapshot.ID, SizeBytes: snapshot.Snapshot.SizeBytes}
	_ = req.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Snapshot %s taken: %s in %s\n", result.SnapshotID,
		unit.DataSize(result.SizeBytes).String(), time.Since(start).Truncate(time.Second)), tasklog.TsNow))
	if err = s.syncSnapshots(ctx, db, target); err != nil {
		// The snapshot is taken: its record comes with the repository's next sync.
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"The repository's snapshot list could not be updated: "+err.Error()+"\n", tasklog.TsNow))
	}
	if req.Sequence == nil {
		req.Task.MustSetOutput(result)
	}
	return &databackupservice.BackupResp{Result: result}, nil
}

// backupCommand streams the command's stdout into the repository. A command that
// fails leaves no snapshot: kopia may have committed what it read, which is
// deleted.
//
// The engine and the command share the run's transaction, which is not safe to
// use from two goroutines: the command starts once the engine is connected, and
// the engine uses the database no more.
func (s *service) backupCommand(
	ctx context.Context,
	db database.Tx,
	req *databackupservice.BackupReq,
	dataBackup *entity.SchedJobDataBackup,
	target backupreposervice.RepoTarget,
	tags []string,
	description string,
) (*backupreposervice.RepoSnapshot, error) {
	reader, writer := io.Pipe()
	streamCtx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()

	type streamed struct {
		resp *backupreposervice.BackupResp
		err  error
		// stoppedByUs is whether the stream ended after the run canceled it: its
		// error then follows from the command's.
		stoppedByUs bool
	}
	connected := make(chan struct{})
	done := make(chan streamed, 1)
	go func() {
		resp, err := s.backupRepoService.BackupStream(streamCtx, db, &backupreposervice.BackupStreamReq{
			RepoTarget: target, Stdin: reader, FileName: dataBackup.SourceFileName, Tags: tags,
			Source:      dataBackupSource + req.JobSetting.ID,
			Description: description,
			OnConnected: func() { close(connected) },
		})
		// Whatever the engine did not read is not waited for.
		_ = reader.CloseWithError(io.ErrClosedPipe)
		done <- streamed{resp: resp, err: err, stoppedByUs: streamCtx.Err() != nil}
	}()

	select {
	case <-connected:
	case result := <-done:
		// The engine stopped before it was connected: the command has not run.
		if result.err == nil {
			return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("the stream ended before it began")
		}
		return nil, hperrors.Wrap(result.err)
	}

	_, execErr := s.schedJobExecService.SchedJobExec(ctx, db, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:    req.TaskExecData,
		SchedJobSetting: req.JobSetting,
		DestApp:         req.App,
		Sequence:        req.Sequence,
		Command:         dataBackup.SourceCommand,
		StdoutWriter:    writer,
	})
	if execErr != nil {
		// The engine must not take a partial output for a whole one.
		cancelStream()
		_ = writer.CloseWithError(execErr)
	} else {
		_ = writer.Close()
	}
	result := <-done

	if execErr != nil {
		s.deleteRunSnapshots(context.WithoutCancel(ctx), db, req, target, result.resp)
		if result.err != nil && !result.stoppedByUs {
			// The engine stopped first, and the command failed writing to it.
			return nil, hperrors.Wrap(result.err)
		}
		return nil, hperrors.Wrap(execErr)
	}
	if result.err != nil {
		return nil, hperrors.Wrap(result.err)
	}
	return result.resp.Snapshot, nil
}

// deleteRunSnapshots deletes what a failed run left in the repository: the
// snapshot the engine named, or those tagged with the run when it named none.
func (s *service) deleteRunSnapshots(
	ctx context.Context,
	db database.IDB,
	req *databackupservice.BackupReq,
	target backupreposervice.RepoTarget,
	resp *backupreposervice.BackupResp,
) {
	var ids []string
	if resp != nil && resp.Snapshot != nil && resp.Snapshot.Snapshot != nil {
		ids = append(ids, resp.Snapshot.Snapshot.ID)
	} else {
		repo, err := target.RepoSetting.AsBackupRepo()
		if err != nil {
			return
		}
		listed, err := s.backupRepoService.ListSnapshots(ctx, db, &backupreposervice.ListSnapshotsReq{
			Scope: target.Scope, Repo: repo, RepoID: target.RepoSetting.ID, RefObjects: target.RefObjects,
			Options: &backup.ListSnapshotsOptions{Tags: []string{dataBackupTagRun + ":" + req.Task.ID}},
		})
		if err != nil {
			_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
				"Could not look for a partial snapshot to delete: "+err.Error()+"\n", tasklog.TsNow))
			return
		}
		for _, item := range listed.Snapshots {
			if item.Snapshot != nil {
				ids = append(ids, item.Snapshot.ID)
			}
		}
	}
	for _, id := range ids {
		err := s.backupRepoService.DeleteSnapshot(ctx, db, &backupreposervice.DeleteSnapshotReq{
			RepoTarget: target, SnapshotID: id,
		})
		if err != nil {
			_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Could not delete the partial snapshot %s: %v\n", id, err), tasklog.TsNow))
			continue
		}
		_ = req.LogStore.Add(ctx, tasklog.NewWarnFrame(
			fmt.Sprintf("Deleted snapshot %s: the command failed, so it holds part of its output\n", id),
			tasklog.TsNow))
	}
}

// backupVolume snapshots the app's part of the volume, on the volume's node.
func (s *service) backupVolume(
	ctx context.Context,
	db database.Tx,
	req *databackupservice.BackupReq,
	dataBackup *entity.SchedJobDataBackup,
	target backupreposervice.RepoTarget,
	tags []string,
	description string,
) (*backupreposervice.RepoSnapshot, error) {
	appVolume, err := s.findAppVolume(ctx, db, req.App, dataBackup.SourceVolume.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dir, err := joinSubpath(appVolume.HostDir, dataBackup.SourceVolumeSubpath)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_ = req.LogStore.Add(ctx, tasklog.NewOutFrame("Backing up "+dir+"\n", tasklog.TsNow))
	resp, err := s.backupRepoService.BackupDirectory(ctx, db, &backupreposervice.BackupDirectoryReq{
		RepoTarget: target, HostDir: dir, NodeID: appVolume.NodeID, NodeLabel: appVolume.NodeLabel, Tags: tags,
		Source: dataBackupSource + req.JobSetting.ID, Description: description,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.Snapshot, nil
}

// syncSnapshots makes the repository's snapshot list match what it now holds.
func (s *service) syncSnapshots(ctx context.Context, db database.Tx, target backupreposervice.RepoTarget) error {
	repo, err := target.RepoSetting.AsBackupRepo()
	if err != nil {
		return hperrors.Wrap(err)
	}
	listed, err := s.backupRepoService.ListSnapshots(ctx, db, &backupreposervice.ListSnapshotsReq{
		Scope: target.Scope, Repo: repo, RepoID: target.RepoSetting.ID, RefObjects: target.RefObjects,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = s.backupRepoService.SyncRepoSnapshots(ctx, db, &backupreposervice.SyncRepoSnapshotsReq{
		Scope: target.Scope, RepoSetting: target.RepoSetting, Remaining: listed.Snapshots,
	})
	return hperrors.Wrap(err)
}
