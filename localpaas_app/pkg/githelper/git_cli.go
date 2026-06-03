package githelper

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

const (
	checkoutMaxDepthDefault = 100
)

type AuthSSHKey struct {
	*ssh.PublicKeys
	PEMBytes []byte
}

type CheckoutOptions struct {
	URL           string
	Auth          transport.AuthMethod
	RemoteName    string
	ReferenceName plumbing.ReferenceName
	SingleBranch  bool
	CommitHash    string

	Depth    uint
	MaxDepth uint

	RecurseSubmodules git.SubmoduleRescursivity
	ShallowSubmodules bool

	LFSEnabled bool

	TempDir     string
	CheckoutDir string
	CacheLoaded bool
	LogStore    *tasklog.Store

	branch    string
	sharedEnv []string
}

func CheckoutWithGitCli(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (repo *git.Repository, commit *object.Commit, err error) {
	// 1. Prepare args
	err = gitCliProcessCheckoutOpts(ctx, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// 2. Clone repository or checkout the latest commit if cache is used
	if checkoutOpts.CacheLoaded {
		repo, err = gitCliCheckoutLatestOnExistingRepo(ctx, checkoutOpts)
	} else {
		repo, err = gitCliClone(ctx, checkoutOpts)
	}
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 3. Checkout target commit
	commit, err = gitCliCheckoutTargetCommit(ctx, repo, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 4. Fetch submodules if needed
	err = gitCliFetchSubmodules(ctx, checkoutOpts)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// 5. Pull LFS files if configured
	// This is done automatically within git clone/pull/fetch commands if GIT_LFS_SKIP_SMUDGE is not set

	return repo, commit, nil
}

func gitCliProcessCheckoutOpts(
	ctx context.Context,
	checkoutOpts *CheckoutOptions,
) (err error) {
	checkoutOpts.sharedEnv = []string{} // No use current process's env
	if !checkoutOpts.LFSEnabled {
		checkoutOpts.sharedEnv = append(checkoutOpts.sharedEnv, "GIT_LFS_SKIP_SMUDGE=1")
	}

	if checkoutOpts.Depth == 0 {
		checkoutOpts.Depth = 1
	}
	checkoutOpts.branch = checkoutOpts.ReferenceName.Short()
	if checkoutOpts.RemoteName == "" {
		checkoutOpts.RemoteName = "origin"
	}

	if checkoutOpts.Auth != nil { //nolint:nestif
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
			// Add user info to the url
			u, err := url.Parse(checkoutOpts.URL)
			if err != nil {
				return apperrors.Wrap(err)
			}
			u.User = url.UserPassword(auth.Username, auth.Password)
			checkoutOpts.URL = u.String()

		case *AuthSSHKey:
			// Use SSH key to clone, the url must be like `git@host.domain:user/repo.git`
			if !strings.HasPrefix(checkoutOpts.URL, "git@") {
				checkoutOpts.URL = GetSshUrl(parseURL)
			}

			sshKeyFile, err := writeSshKeyFile(checkoutOpts.TempDir, auth.PEMBytes)
			if err != nil {
				addLog(ctx, fmt.Sprintf("Failed to write SSH key file: %v error: %v",
					sshKeyFile, err.Error()), true, checkoutOpts)
				return apperrors.Wrap(err)
			}
			sshCmd := "ssh -o StrictHostKeyChecking=no -i " + sshKeyFile
			checkoutOpts.sharedEnv = append(checkoutOpts.sharedEnv, "GIT_SSH_COMMAND="+sshCmd)

		default:
			addLog(ctx, fmt.Sprintf("Git auth method '%v' is unsupported", auth.Name()),
				true, checkoutOpts)
			return apperrors.NewUnsupported(apperrors.Fmt("Git auth method '%v'", auth.Name()))
		}
	}

	return nil
}
