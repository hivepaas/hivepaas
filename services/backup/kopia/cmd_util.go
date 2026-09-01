package kopia

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const (
	cmdSnapshot   = "snapshot"
	cmdRepository = "repository"

	cmdCreate = "create"
	cmdList   = "list"

	cmdFlagJSON = "--json"
)

func (c *Client) buildStorageFlags() ([]string, error) {
	if c.storage == nil {
		return nil, hperrors.Wrap(backupmodel.ErrStorageConfigRequired)
	}
	if c.storage.StorageS3 != nil {
		return c.buildS3Flags(c.storage.StorageS3), nil
	}
	if c.storage.StorageLocal != nil {
		return c.buildLocalFlags(c.storage.StorageLocal), nil
	}
	return nil, hperrors.Wrap(backupmodel.ErrStorageTypeRequired)
}

func (c *Client) buildS3Flags(cfg *backupmodel.StorageS3) []string {
	// NOTE: credentials are deliberately NOT passed as flags. They go through the environment
	// (see buildEnv), so they stay out of argv - out of `ps`, out of error messages, out of logs.
	flags := []string{
		"s3",
		"--bucket=" + cfg.Bucket,
	}

	if cfg.Endpoint != "" {
		endpoint := strings.TrimSpace(cfg.Endpoint)
		// kopia wants a bare host and assumes HTTPS. A self-hosted endpoint given as http:// has
		// to say so explicitly, otherwise kopia speaks TLS to a plaintext port and fails.
		plaintext := strings.HasPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
		flags = append(flags, "--endpoint="+endpoint)
		if plaintext {
			flags = append(flags, "--disable-tls")
		}
	}

	if cfg.Region != "" {
		flags = append(flags, "--region="+cfg.Region)
	}

	if cfg.Prefix != "" {
		prefix := strings.Trim(cfg.Prefix, "/") + "/"
		flags = append(flags, "--prefix="+prefix)
	}

	return flags
}

func (c *Client) buildLocalFlags(cfg *backupmodel.StorageLocal) []string {
	return []string{
		"filesystem",
		"--path=" + cfg.Path,
	}
}

// buildGlobalFlags returns the kopia global flags that must precede every subcommand.
func (c *Client) buildGlobalFlags() []string {
	if c.storage == nil || c.storage.ConfigFile == "" {
		return nil
	}
	return []string{"--config-file=" + c.storage.ConfigFile}
}

// Environment variables kopia reads its secrets from. Passing secrets this way keeps them out of
// argv, where they would otherwise be visible to `ps` and would end up in error messages.
const (
	envPassword    = "KOPIA_PASSWORD"
	envNewPassword = "KOPIA_NEW_PASSWORD" //nolint:gosec // variable name, not a credential
	envS3AccessKey = "AWS_ACCESS_KEY_ID"
	envS3SecretKey = "AWS_SECRET_ACCESS_KEY" //nolint:gosec // variable name, not a credential
)

func (c *Client) buildEnv(password string) []string {
	pwd := password
	if pwd == "" && c.storage != nil {
		pwd = c.storage.RepositoryPassword
	}

	env := []string{
		envPassword + "=" + pwd,
	}

	if c.storage != nil && c.storage.StorageS3 != nil {
		if c.storage.StorageS3.AccessKey != "" {
			env = append(env, envS3AccessKey+"="+c.storage.StorageS3.AccessKey)
		}
		if c.storage.StorageS3.SecretKey != "" {
			env = append(env, envS3SecretKey+"="+c.storage.StorageS3.SecretKey)
		}
	}

	return env
}

type execOptions struct {
	env       []string
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	nodeID    string
	nodeLabel string
}

//nolint:unparam
func (c *Client) execCommand(
	ctx context.Context,
	args []string,
	opts ...func(*execOptions),
) (resp *backupmodel.CommandExecResp, err error) {
	if c.commandExec == nil {
		return nil, hperrors.Wrap(backupmodel.ErrCommandExecutorMissing)
	}

	nodeID, nodeLabel := c.getNodeInfo()
	opt := &execOptions{
		env:       c.buildEnv(""),
		nodeID:    nodeID,
		nodeLabel: nodeLabel,
	}
	for _, fn := range opts {
		fn(opt)
	}

	fullCommand := append([]string{"kopia"}, c.buildGlobalFlags()...)
	fullCommand = append(fullCommand, args...)

	req := &backupmodel.CommandExecReq{
		Command:   fullCommand,
		Env:       opt.env,
		Stdin:     opt.stdin,
		Stdout:    opt.stdout,
		Stderr:    opt.stderr,
		NodeID:    opt.nodeID,
		NodeLabel: opt.nodeLabel,
	}

	resp, err = c.commandExec(ctx, req)
	if err != nil {
		return resp, hperrors.Wrap(fmt.Errorf("%w: %s: %w",
			backupmodel.ErrCommandFailed, redactCommand(fullCommand), err))
	}
	if resp != nil && resp.ExitCode != 0 {
		return resp, hperrors.Wrap(fmt.Errorf("%w: %s: exit code %d",
			backupmodel.ErrCommandFailed, redactCommand(fullCommand), resp.ExitCode))
	}

	return resp, nil
}

// secretFlags are the flags whose value must never end up in an error message or a log: errors
// built here travel all the way to the API response.
var secretFlags = []string{
	"--access-key",
	"--secret-access-key",
	"--password",
	"--new-password",
}

// redactCommand renders a command for humans with its credentials masked.
func redactCommand(command []string) string {
	parts := make([]string, 0, len(command))
	for _, arg := range command {
		parts = append(parts, redactArg(arg))
	}
	return strings.Join(parts, " ")
}

func redactArg(arg string) string {
	name, _, found := strings.Cut(arg, "=")
	if !found {
		return arg
	}
	for _, flag := range secretFlags {
		if name == flag {
			return name + "=***"
		}
	}
	return arg
}

func (c *Client) getNodeInfo() (nodeID, nodeLabel string) {
	if c.storage != nil && c.storage.StorageLocal != nil {
		return c.storage.StorageLocal.NodeID, c.storage.StorageLocal.NodeLabel
	}
	return "", ""
}
