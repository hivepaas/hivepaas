package gittool

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

type CheckoutOptions struct {
	URL         string
	Credentials *entity.Setting

	RemoteName    string
	ReferenceName plumbing.ReferenceName
	CommitHash    string

	SubmodulesEnabled bool
	LFSEnabled        bool

	TempDir     string
	CheckoutDir string
	CacheLoaded bool
	LogStore    *tasklog.Store

	// Internal fields
	refType  githelper.RefType
	refShort string
}

func CheckoutWithGitCli(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (repo *git.Repository, commit *object.Commit, err error) {
	cli := &checkoutCli{
		opts: checkoutOpts,
	}
	return cli.checkout(ctx)
}

type checkoutCli struct {
	opts        *CheckoutOptions
	sharedEnv   []string
	needCleanup bool
}

func (cli *checkoutCli) checkout(
	ctx context.Context,
) (repo *git.Repository, commit *object.Commit, err error) {
	// 1. Prepare args
	if err = cli.processCheckoutOpts(ctx); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 2. Clone repository if source cache is not there
	if !cli.opts.CacheLoaded {
		if err = cli.clone(ctx); err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Open repo with go-git
	if repo, err = git.PlainOpen(cli.opts.CheckoutDir); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 3. Checkout target commit
	if commit, err = cli.checkoutTargetCommit(ctx, repo); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 4. Fetch submodules if needed
	if err = cli.fetchSubmodules(ctx); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 5. Pull LFS files if configured
	// This is done automatically within git clone/pull/fetch commands if GIT_LFS_SKIP_SMUDGE is not set

	// 6. Cleanup orphaned data
	if cli.needCleanup {
		if err = cli.cleanup(ctx); err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
	}

	return repo, commit, nil
}

func (cli *checkoutCli) processCheckoutOpts(
	ctx context.Context,
) (err error) {
	cli.sharedEnv = []string{} // No use current process's env
	if !cli.opts.LFSEnabled {
		cli.sharedEnv = append(cli.sharedEnv, "GIT_LFS_SKIP_SMUDGE=1")
	}

	if cli.opts.RemoteName == "" {
		cli.opts.RemoteName = "origin"
	}

	cli.opts.refType, cli.opts.refShort = githelper.GetRefShort(string(cli.opts.ReferenceName))
	if !cli.opts.refType.CanCheckout() {
		return apperrors.NewUnsupported("Repository ref type")
	}

	authMethod, err := calcGitAuthMethod(ctx, cli.opts.Credentials)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if authMethod != nil { //nolint:nestif
		parseURL, err := vcsurl.Parse(cli.opts.URL)
		if err != nil {
			return apperrors.Wrap(err)
		}

		switch auth := authMethod.(type) {
		case *http.BasicAuth:
			// Use https url
			if !strings.HasPrefix(cli.opts.URL, "https://") {
				cli.opts.URL = githelper.GetHttpsUrl(parseURL)
			}
			// Add user info to the url
			u, err := url.Parse(cli.opts.URL)
			if err != nil {
				return apperrors.Wrap(err)
			}
			u.User = url.UserPassword(auth.Username, auth.Password)
			cli.opts.URL = u.String()

		case *authSSHKey:
			// Use SSH key to clone, the url must be like `git@host.domain:user/repo.git`
			if !strings.HasPrefix(cli.opts.URL, "git@") {
				cli.opts.URL = githelper.GetSshUrl(parseURL)
			}

			sshKeyFile, err := writeSshKeyFile(cli.opts.TempDir, auth.PEMBytes)
			if err != nil {
				addLog(ctx, fmt.Sprintf("Failed to write SSH key file: %v error: %v",
					sshKeyFile, err.Error()), true, cli.opts.LogStore)
				return apperrors.Wrap(err)
			}
			sshCmd := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i " + sshKeyFile
			cli.sharedEnv = append(cli.sharedEnv, "GIT_SSH_COMMAND="+sshCmd)

		default:
			addLog(ctx, fmt.Sprintf("Git auth method '%v' is unsupported", auth.Name()),
				true, cli.opts.LogStore)
			return apperrors.Wrap(apperrors.ErrGitAuthMethodUnsupported).WithParam("AuthMethod", auth.Name())
		}
	}

	return nil
}
