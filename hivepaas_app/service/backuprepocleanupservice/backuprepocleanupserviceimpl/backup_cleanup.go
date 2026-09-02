package backuprepocleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dblock"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backuprepocleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
)

type cleanupData struct {
	*backuprepocleanupservice.BackupRepoCleanupReq
	TaskOutput *entity.TaskBackupRepoCleanupOutput
}

func (s *service) Cleanup(
	ctx context.Context,
	db database.Tx,
	req *backuprepocleanupservice.BackupRepoCleanupReq,
) (resp *backuprepocleanupservice.BackupRepoCleanupResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	data := &cleanupData{
		BackupRepoCleanupReq: req,
		TaskOutput:           &entity.TaskBackupRepoCleanupOutput{},
	}
	if data.LogStore == nil {
		data.LogStore = tasklog.NewLocalStore(fmt.Sprintf("task:%v:log", req.Task.ID))
	}
	resp = &backuprepocleanupservice.BackupRepoCleanupResp{
		TaskOutput: data.TaskOutput,
	}

	repos, err := s.loadTargetRepos(ctx, db, data)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	if len(repos) == 0 {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("No backup repository to clean up", nil))
		data.Task.MustSetOutput(data.TaskOutput)
		return resp, nil
	}

	// One repository failing must not stop the rest: they are independent, and a repository whose
	// storage is unreachable would otherwise block every other one from ever being pruned.
	var errs []error
	for _, repo := range repos {
		if err := s.cleanupOneRepo(ctx, db, data, repo); err != nil {
			errs = append(errs, err)
		}
	}

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	return resp, errors.Join(errs...)
}

// loadTargetRepos resolves what to clean: the repositories the task named, or every active one
// when it named none.
func (s *service) loadTargetRepos(
	ctx context.Context,
	db database.IDB,
	data *cleanupData,
) ([]*entity.Setting, error) {
	args, err := data.Task.ArgsAsBackupRepoCleanup()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	}
	if args != nil && len(args.TargetRepos) > 0 {
		opts = append(opts, bunex.SelectWhereIn("setting.id IN (?)", args.TargetRepos.ToIDStringSlice()...))
	}

	// A nil scope means no scope filter, so this covers repositories in every scope - global,
	// project, or otherwise. There is no single scope that would make sense here: the job cleans
	// all of them, and each repository's own scope is resolved per repository below.
	repos, _, err := s.settingRepo.List(ctx, db, nil, nil,
		append(opts, bunex.SelectWhere("setting.type = ?", base.SettingTypeBackupRepo))...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return repos, nil
}

// cleanupOneRepo prunes a single repository and reconciles its snapshot records.
func (s *service) cleanupOneRepo(
	ctx context.Context,
	db database.Tx,
	data *cleanupData,
	repoSetting *entity.Setting,
) error {
	output := &entity.TaskBackupRepoCleanupRepoOutput{
		RepoID:   repoSetting.ID,
		RepoName: repoSetting.Name,
	}
	data.TaskOutput.Repos = append(data.TaskOutput.Repos, output)

	// Downstream work has to run in the repository's own scope, not the job's. Resolving the
	// storage credentials goes through the scope filter, which is what lets a project repository
	// reach a cloud storage that is inherited from global or shared into the project - a nil or
	// global scope would not find it.
	//
	// A repository whose project is gone is skipped rather than failed: there is nothing to
	// clean up, and one dead project should not fail the whole run.
	scope, err := s.scopeService.LoadObjectScope(ctx, db, repoSetting.Scope, repoSetting.ObjectID, true)
	if err != nil {
		output.Skipped = true
		data.TaskOutput.ReposSkipped++
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Skipped '"+repoSetting.Name+"': its scope is no longer available", nil))
		return nil
	}

	// The same lock the manual cleanup endpoint takes, so a scheduled run and a user-triggered
	// one can never prune the same repository at once. Being refused is a normal outcome here:
	// the other run is doing the work, so this one moves on to the next repository.
	lock, acquired, err := dblock.TryAcquire(ctx, s.db, backupreposervice.RepoLockName(repoSetting.ID))
	if err != nil {
		output.Error = err.Error()
		data.TaskOutput.ReposFailed++
		return hperrors.Wrap(err)
	}
	if !acquired {
		output.Skipped = true
		data.TaskOutput.ReposSkipped++
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Skipped '"+repoSetting.Name+"': a cleanup is already running for it", nil))
		return nil
	}
	defer func() {
		_ = lock.Release(ctx)
	}()

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Cleaning up backup repository: "+repoSetting.Name, nil))

	cleanupResp, err := s.backupRepoService.CleanupRepo(ctx, db, &backupreposervice.CleanupRepoReq{
		Scope:       scope,
		RepoSetting: repoSetting,
	})
	if err != nil {
		output.Error = hperrors.GetErrorDetail(err, "")
		data.TaskOutput.ReposFailed++
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Failed to clean up '"+repoSetting.Name+"': "+err.Error(), nil))
		return hperrors.Wrap(err)
	}

	syncResp, err := s.backupRepoService.SyncRepoSnapshots(ctx, db, &backupreposervice.SyncRepoSnapshotsReq{
		Scope:       scope,
		RepoSetting: repoSetting,
		Remaining:   cleanupResp.Remaining,
	})
	if err != nil {
		output.Error = hperrors.GetErrorDetail(err, "")
		data.TaskOutput.ReposFailed++
		return hperrors.Wrap(err)
	}

	output.SnapshotsInRepo = len(cleanupResp.Remaining)
	output.RecordsRemoved = len(syncResp.Removed)
	output.RecordsAdded = syncResp.Added
	data.TaskOutput.ReposCleaned++

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf(
		"Cleaned up '%s': %d snapshot(s) left, %d record(s) removed, %d added",
		repoSetting.Name, output.SnapshotsInRepo, output.RecordsRemoved, output.RecordsAdded), nil))

	return nil
}
