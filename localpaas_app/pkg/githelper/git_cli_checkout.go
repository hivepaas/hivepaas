package githelper

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
)

func gitCliCheckoutLatestOnExistingRepo(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (repo *git.Repository, err error) {
	// Fetch latest commit
	fetchArgs := []string{"fetch", "--depth", "1", checkoutOpts.RemoteName, checkoutOpts.branch}

	fetchCmd := exec.CommandContext(ctx, "git", fetchArgs...)
	fetchCmd.Dir = checkoutOpts.CheckoutDir
	fetchCmd.Env = checkoutOpts.sharedEnv

	out, err := fetchCmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, checkoutOpts)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Hard reset the branch to make it point to the last fetched commit
	resetArgs := []string{"reset", "--hard", fmt.Sprintf("%s/%s", checkoutOpts.RemoteName, checkoutOpts.branch)}

	resetCmd := exec.CommandContext(ctx, "git", resetArgs...)
	resetCmd.Dir = checkoutOpts.CheckoutDir
	resetCmd.Env = checkoutOpts.sharedEnv

	out, err = resetCmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, checkoutOpts)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// git gc --prune=now
	pruneCmd := exec.CommandContext(ctx, "git", "gc", "--prune=now")
	pruneCmd.Dir = checkoutOpts.CheckoutDir
	pruneCmd.Env = checkoutOpts.sharedEnv

	out, err = pruneCmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, checkoutOpts)

	// Open repo with go-git
	repo, err = git.PlainOpen(checkoutOpts.CheckoutDir)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return repo, nil
}

func gitCliCheckoutTargetCommit(
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
	// Try to resolve target commit
	commit, err = repo.CommitObject(targetHash)

	if err != nil && errors.Is(err, plumbing.ErrObjectNotFound) {
		// Need to fetch deeper
		depth := uint(0)
		maxDepth := gofn.Coalesce(checkoutOpts.MaxDepth, checkoutMaxDepthDefault)
		depthIncrement := max(20, maxDepth/10) //nolint:mnd

		for depth <= maxDepth {
			depth += depthIncrement
			fetchArgs := []string{"fetch", "origin", "--depth", strconv.FormatUint(uint64(depth), 10)}
			if checkoutOpts.branch != "" {
				fetchArgs = append(fetchArgs, checkoutOpts.branch)
			}

			fetchCmd := exec.CommandContext(ctx, "git", fetchArgs...)
			fetchCmd.Dir = checkoutOpts.CheckoutDir
			fetchCmd.Env = checkoutOpts.sharedEnv

			out, err := fetchCmd.CombinedOutput()
			addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, checkoutOpts)
			if err != nil {
				return nil, apperrors.Wrap(err)
			}

			commit, err = repo.CommitObject(targetHash)
			if err == nil && commit != nil {
				break
			}
		}
	}

	if commit == nil {
		addLog(ctx, fmt.Sprintf("Failed to checkout commit: %v, commit is too deep or doesn't exist.",
			checkoutOpts.CommitHash), err != nil, checkoutOpts)
		return nil, apperrors.NewNotFound(apperrors.Fmt("Commit '%v'", checkoutOpts.CommitHash))
	}

	// Checkout target commit
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", checkoutOpts.CommitHash) //nolint:gosec
	checkoutCmd.Dir = checkoutOpts.CheckoutDir
	checkoutCmd.Env = checkoutOpts.sharedEnv

	out, err := checkoutCmd.CombinedOutput()
	addLog(ctx, reflectutil.UnsafeBytesToStr(out), err != nil, checkoutOpts)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return commit, nil
}
