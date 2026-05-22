package githelper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func CheckoutWithGoGit(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (repo *git.Repository, commit *object.Commit, err error) {
	// 1. Prepare args
	err = goGitProcessCheckoutOpts(ctx, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// 2. Clone repository using go-git
	repo, err = goGitClone(ctx, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 3. Checkout target commit
	commit, err = goGitCheckoutTargetCommit(ctx, repo, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return repo, commit, nil
}

func goGitProcessCheckoutOpts(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (err error) {
	if checkoutOpts.Auth != nil {
		parseURL, err := vcsurl.Parse(checkoutOpts.URL)
		if err != nil {
			return apperrors.Wrap(err)
		}

		switch auth := checkoutOpts.Auth.(type) {
		case *http.BasicAuth:
			// Use https url
			if !strings.HasPrefix(checkoutOpts.URL, "https://") {
				checkoutOpts.URL = GetHttpsUrl(parseURL)
			}

		case *AuthSSHKey:
			// Use SSH key to clone, the url must be like `git@host.domain:user/repo.git`
			if !strings.HasPrefix(checkoutOpts.URL, "git@") {
				checkoutOpts.URL = GetSshUrl(parseURL)
			}

		default:
			addLog(ctx, fmt.Sprintf("Git auth method '%v' is unsupported", auth.Name()),
				true, checkoutOpts)
			return apperrors.NewUnsupported(fmt.Sprintf("Git auth method '%v'", auth.Name()))
		}
	}

	if checkoutOpts.Depth == 0 {
		checkoutOpts.Depth = 1
	}
	checkoutOpts.branch = checkoutOpts.ReferenceName.Short()

	return nil
}

func goGitClone(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (repo *git.Repository, err error) {
	err = os.MkdirAll(checkoutOpts.CheckoutDir, checkoutPathFileMode)
	if err != nil {
		return nil, apperrors.New(err)
	}

	repo, err = git.PlainCloneContext(ctx, checkoutOpts.CheckoutDir, false, &git.CloneOptions{
		URL:               checkoutOpts.URL,
		Auth:              checkoutOpts.Auth,
		RemoteName:        checkoutOpts.RemoteName,
		ReferenceName:     checkoutOpts.ReferenceName,
		SingleBranch:      checkoutOpts.SingleBranch,
		Depth:             int(checkoutOpts.Depth), //nolint
		RecurseSubmodules: checkoutOpts.RecurseSubmodules,
		ShallowSubmodules: checkoutOpts.ShallowSubmodules,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return repo, nil
}

func goGitCheckoutTargetCommit(
	ctx context.Context,
	repo *git.Repository,
	checkoutOpts *CheckoutOptions,
) (commit *object.Commit, err error) {
	if checkoutOpts.CommitHash == "" {
		head, err := repo.Head()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		commit, err = repo.CommitObject(head.Hash())
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return commit, nil
	}

	targetHash := plumbing.NewHash(checkoutOpts.CommitHash)
	depth := uint(0)
	maxDepth := gofn.Coalesce(checkoutOpts.MaxDepth, checkoutMaxDepthDefault)
	depthIncrement := max(20, maxDepth/10) //nolint:mnd

	for depth <= maxDepth {
		commit, err = repo.CommitObject(targetHash)
		if err == nil && commit != nil {
			break
		}
		if !errors.Is(err, plumbing.ErrObjectNotFound) {
			return nil, apperrors.Wrap(err)
		}
		depth += depthIncrement
		err = repo.FetchContext(ctx, &git.FetchOptions{
			RemoteName: git.DefaultRemoteName,
			Depth:      int(depth), //nolint
			Auth:       checkoutOpts.Auth,
			Force:      true,
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	if commit == nil {
		addLog(ctx, fmt.Sprintf("Failed to checkout commit: %v, commit is too deep or doesn't exist.",
			checkoutOpts.CommitHash), err != nil, checkoutOpts)
		return nil, apperrors.NewNotFound(fmt.Sprintf("Commit '%v'", checkoutOpts.CommitHash))
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: targetHash,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return commit, nil
}
