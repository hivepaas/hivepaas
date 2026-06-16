package gittool

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
)

type ValidationOptions struct {
	URL         string
	Credentials *entity.Setting

	RemoteName    string
	ReferenceName plumbing.ReferenceName
	CommitHash    string
	TempDir       string
}

func ValidateWithGitCli(
	ctx context.Context,
	checkoutOpts *ValidationOptions,
) (err error) {
	cli := &validationCli{
		opts: checkoutOpts,
	}
	return cli.validate(ctx)
}

type validationCli struct {
	opts           *ValidationOptions
	sharedEnv      []string
	cleanupTempDir bool

	refShort    string
	refIsBranch bool
}

func (cli *validationCli) validate(
	ctx context.Context,
) (err error) {
	// 1. Prepare args
	if err = cli.processValidationOpts(ctx); err != nil {
		return apperrors.New(err)
	}
	defer cli.cleanup()

	// 2. Validate repository URL
	cmd := exec.CommandContext(ctx, "git", "ls-remote", cli.opts.URL) //nolint:gosec
	cmd.Env = cli.sharedEnv
	out, err := cmd.CombinedOutput()
	if err != nil {
		return apperrors.NewNotFound("Repository")
	}

	// 3. Validate reference
	refsMap := make(map[string]bool)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 { //nolint:mnd
			refsMap[parts[1]] = true
		}
	}

	if cli.refIsBranch {
		if !refsMap["refs/heads/"+cli.refShort] {
			return apperrors.NewNotFound("Repository branch")
		}
	} else {
		if !refsMap["refs/tags/"+cli.refShort] {
			return apperrors.NewNotFound("Repository branch")
		}
	}

	// 4. Validate commit hash
	if err = cli.validateCommit(ctx); err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (cli *validationCli) validateCommit(
	ctx context.Context,
) (err error) {
	if cli.opts.CommitHash == "" {
		return nil
	}
	if !githelper.IsCommitHash(cli.opts.CommitHash) {
		return apperrors.NewParamInvalid("Repository commit hash")
	}

	cloneDir, err := os.MkdirTemp(cli.opts.TempDir, "git-clone-*")
	if err != nil {
		return apperrors.New(err)
	}
	defer os.RemoveAll(cloneDir)

	//nolint:gosec
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--filter=blob:none",
		"--single-branch", "--branch", cli.refShort, "--", cli.opts.URL, cloneDir)
	cloneCmd.Env = cli.sharedEnv
	if cloneOut, err := cloneCmd.CombinedOutput(); err != nil {
		return apperrors.Wrap(fmt.Errorf("clone failed: %w (output: %s)", err, string(cloneOut)))
	}

	//nolint:gosec
	catCmd := exec.CommandContext(ctx, "git", "cat-file", "-e", cli.opts.CommitHash+"^{commit}")
	catCmd.Dir = cloneDir
	catCmd.Env = []string{}
	if err := catCmd.Run(); err == nil {
		//nolint:gosec
		ancestorCmd := exec.CommandContext(ctx, "git", "merge-base", "--is-ancestor",
			cli.opts.CommitHash, "HEAD")
		ancestorCmd.Dir = cloneDir
		ancestorCmd.Env = []string{}
		if err := ancestorCmd.Run(); err == nil {
			return nil
		}
	}

	return apperrors.NewNotFound("Repository commit")
}

func (cli *validationCli) processValidationOpts(
	ctx context.Context,
) (err error) {
	// Creates temp dir if empty
	if cli.opts.TempDir == "" {
		cli.opts.TempDir, err = fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)
		if err != nil {
			return apperrors.Wrap(err)
		}
		cli.cleanupTempDir = true
	}

	cli.sharedEnv = []string{} // No use current process's env
	if cli.opts.RemoteName == "" {
		cli.opts.RemoteName = "origin"
	}

	cli.refIsBranch = cli.opts.ReferenceName.IsBranch()
	if cli.refIsBranch || cli.opts.ReferenceName.IsTag() { // Only support branch and tag
		cli.refShort = cli.opts.ReferenceName.Short()
	} else {
		return apperrors.NewUnsupported("Repository branch")
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
				return apperrors.Wrap(err)
			}
			sshCmd := "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i " + sshKeyFile
			cli.sharedEnv = append(cli.sharedEnv, "GIT_SSH_COMMAND="+sshCmd)

		default:
			return apperrors.NewUnsupported(apperrors.Fmt("Git auth method '%v'", auth.Name()))
		}
	}

	return nil
}

func (cli *validationCli) cleanup() {
	if cli.cleanupTempDir {
		_ = os.RemoveAll(cli.opts.TempDir)
	}
}
