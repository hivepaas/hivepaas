package repocheckoutserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/gittool"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
)

type repoCheckoutData struct {
	*repocheckoutservice.RepoCheckoutReq
	Resp *repocheckoutservice.RepoCheckoutResp

	RepoCacheFile    *entity.File
	RepoCacheLoaded  bool
	CheckoutDuration time.Duration
}

func (s *service) Checkout(
	ctx context.Context,
	req *repocheckoutservice.RepoCheckoutReq,
) (resp *repocheckoutservice.RepoCheckoutResp, err error) {
	resp = &repocheckoutservice.RepoCheckoutResp{}
	data := &repoCheckoutData{
		RepoCheckoutReq: req,
		Resp:            resp,
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
		}
	}()

	err = s.doCheckout(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, err
}

func (s *service) doCheckout(
	ctx context.Context,
	data *repoCheckoutData,
) (err error) {
	repoSource := data.RepoSource

	err = s.checkoutPrepare(data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// NOTE: currently supports repo of git type only
	if repoSource.RepoType != base.RepoTypeGit {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to checkout source: "+
			"unsupported repository type: "+string(repoSource.RepoType), tasklog.TsNow))
		return apperrors.Wrap(apperrors.ErrRepoTypeUnsupported).WithParam("Type", repoSource.RepoType)
	}

	s.addStepStartLog(ctx, data, "Start cloning Git repository...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	err = s.loadRepoCache(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return apperrors.Wrap(err)
	}

	checkoutOptions := &gittool.CheckoutOptions{
		URL:         repoSource.RepoURL,
		Credentials: data.CredSetting,

		ReferenceName:     plumbing.ReferenceName(repoSource.RepoRef),
		CommitHash:        repoSource.CommitHash,
		SubmodulesEnabled: repoSource.RepoOptions.GitSubmodulesEnabled,
		LFSEnabled:        repoSource.RepoOptions.GitLFSEnabled,

		TempDir:     data.TempDir,
		CheckoutDir: data.CheckoutDir,
		CacheLoaded: data.RepoCacheLoaded,
		LogStore:    data.LogStore,
	}

	var commit *object.Commit
	checkoutStart := time.Now()
	for {
		_, commit, err = gittool.CheckoutWithGitCli(ctx, checkoutOptions)
		if err == nil {
			break
		}
		if checkoutOptions.CacheLoaded {
			if err := s.resetCheckoutDir(data); err != nil {
				return apperrors.Wrap(err)
			}
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Failed to checkout repository with error: "+
				err.Error()+". Try to do a fresh clone (not using cache)...", tasklog.TsNow))
			checkoutOptions.CacheLoaded = false
			data.RepoCacheLoaded = false
			continue
		}
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to checkout repository with error: "+
			err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	data.CheckoutDuration = time.Since(checkoutStart)
	data.Resp.CommitHash = commit.Hash.String()
	data.Resp.CommitMessage = commit.Message
	data.Resp.CommitTitle = strutil.GetFirstLine(commit.Message)
	data.Resp.CommitAuthor = commit.Author.Name

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return apperrors.Wrap(err)
	}

	// Cache the latest repo source if satisfied our condition
	ee := s.saveRepoCache(ctx, data)
	if ee != nil { // Just log
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to cache repository source: "+
			ee.Error(), tasklog.TsNow))
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return apperrors.Wrap(err)
	}

	// Remove .git dir within the source dir before building image
	ee = os.RemoveAll(filepath.Join(data.CheckoutDir, ".git"))
	if ee != nil { // Just log
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to remove .git directory from source: "+
			ee.Error(), tasklog.TsNow))
	}

	return nil
}

func (s *service) checkoutPrepare(
	data *repoCheckoutData,
) (err error) {
	repoSource := data.RepoSource

	// Loads repo credentials (github app, git token, ssh key) if configured
	if repoSource.Credentials.ID != "" {
		data.CredSetting = data.RefObjects.RefSettings[repoSource.Credentials.ID]
	}

	// Creates checkout dir
	if data.CheckoutDir == "" {
		data.CheckoutDir = filepath.Join(data.TempDir, "checkout")
	}
	err = os.MkdirAll(data.CheckoutDir, base.DirModeDefault)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) addStepStartLog(
	ctx context.Context,
	data *repoCheckoutData,
	msg string,
) {
	_ = data.LogStore.Add(ctx,
		tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
		tasklog.NewOutFrame(msg, tasklog.TsNow))
}

func (s *service) addStepEndLog(
	ctx context.Context,
	data *repoCheckoutData,
	start time.Time,
	err error,
) {
	duration := timeutil.NowUTC().Sub(start).Truncate(time.Millisecond)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String()+
			" with error: "+err.Error(), tasklog.TsNow))
	} else {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String(),
			tasklog.TsNow))
	}
}
