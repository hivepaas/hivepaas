package kopia

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// noUpdateCheckFlag disables kopia's periodic GitHub update check. Nodes are not guaranteed
// to have outbound internet access, and the check only adds latency to every repo operation.
const noUpdateCheckFlag = "--no-check-for-updates"

func (c *Client) InitRepo(
	ctx context.Context,
	opts *backupmodel.InitRepoOptions,
) error {
	storageFlags, err := c.buildStorageFlags()
	if err != nil {
		return hperrors.Wrap(err)
	}

	args := append([]string{cmdRepository, cmdCreate}, storageFlags...)
	args = append(args, noUpdateCheckFlag)
	if opts != nil && opts.Description != "" {
		args = append(args, "--description="+opts.Description)
	}

	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, args, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return hperrors.Wrap(fmt.Errorf("kopia repository create failed: %s (err: %w)", errMsg, err))
		}
		return hperrors.Wrap(fmt.Errorf("kopia repository create failed: %w", err))
	}

	if err := c.applyRepoOptions(ctx, opts); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// applyRepoOptions applies the settings that cannot be passed to `repository create` directly.
// The repository already exists at this point, so a failure here leaves a usable repository
// running with kopia defaults instead of the requested values.
func (c *Client) applyRepoOptions(
	ctx context.Context,
	opts *backupmodel.InitRepoOptions,
) error {
	if opts == nil {
		return nil
	}

	if opts.PackSizeMB > 0 {
		var errBuf bytes.Buffer
		_, err := c.execCommand(ctx,
			[]string{cmdRepository, "set-parameters", "--max-pack-size-mb=" + strconv.Itoa(opts.PackSizeMB)},
			func(o *execOptions) { o.stderr = &errBuf })
		if err != nil {
			return hperrors.Wrap(fmt.Errorf("kopia repository set-parameters failed: %s (err: %w)",
				strings.TrimSpace(errBuf.String()), err))
		}
	}

	if opts.Compression != "" {
		var errBuf bytes.Buffer
		_, err := c.execCommand(ctx,
			[]string{"policy", "set", "--global", "--compression=" + opts.Compression},
			func(o *execOptions) { o.stderr = &errBuf })
		if err != nil {
			return hperrors.Wrap(fmt.Errorf("kopia policy set compression failed: %s (err: %w)",
				strings.TrimSpace(errBuf.String()), err))
		}
	}

	return nil
}

func (c *Client) ConnectRepo(
	ctx context.Context,
) error {
	storageFlags, err := c.buildStorageFlags()
	if err != nil {
		return hperrors.Wrap(err)
	}

	args := append([]string{cmdRepository, "connect"}, storageFlags...)
	args = append(args, noUpdateCheckFlag)

	var errBuf bytes.Buffer
	_, err = c.execCommand(ctx, args, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return hperrors.Wrap(fmt.Errorf("kopia repository connect failed: %s (err: %w)", errMsg, err))
		}
		return hperrors.Wrap(fmt.Errorf("kopia repository connect failed: %w", err))
	}
	return nil
}

func (c *Client) CheckRepo(
	ctx context.Context,
) error {
	var errBuf bytes.Buffer
	_, err := c.execCommand(ctx, []string{cmdRepository, "validate-provider"}, func(o *execOptions) {
		o.stderr = &errBuf
	})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return hperrors.Wrap(fmt.Errorf("kopia check repository failed: %s (err: %w)", errMsg, err))
		}
		return hperrors.Wrap(fmt.Errorf("kopia check repository failed: %w", err))
	}
	return nil
}

func (c *Client) ChangePassword(
	ctx context.Context,
	oldPassword, newPassword string,
) error {
	var errBuf bytes.Buffer
	// Both passwords travel through the environment rather than argv.
	_, err := c.execCommand(ctx, []string{cmdRepository, "change-password"},
		func(o *execOptions) {
			o.env = append(c.buildEnv(oldPassword), envNewPassword+"="+newPassword)
			o.stderr = &errBuf
		})
	if err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		if errMsg != "" {
			return hperrors.Wrap(fmt.Errorf("kopia change password failed: %s (err: %w)", errMsg, err))
		}
		return hperrors.Wrap(fmt.Errorf("kopia change password failed: %w", err))
	}

	if c.storage != nil {
		c.storage.RepositoryPassword = newPassword
	}
	return nil
}
