package kopia

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const cmdMaintenance = "maintenance"

// takeMaintenance makes this client the repository's maintenance owner, the one client kopia
// runs maintenance for: the one that created the repository, on whatever host that was, or the
// last to take it. HivePaaS connects as one identity wherever its commands run
// (Storage.Identity); a repository whose owner is another - an earlier container's hostname, the
// machine it was created on - is taken over, as kopia's own `maintenance set --owner=me` does.
// It gives the owner it was taken from, empty when it was this client already.
func (c *Client) takeMaintenance(ctx context.Context) (string, error) {
	var info struct {
		Owner string `json:"owner"`
	}
	if err := c.readJSON(ctx, []string{cmdMaintenance, "info", cmdFlagJSON}, &info); err != nil {
		return "", hperrors.Wrap(err)
	}
	var status struct {
		ClientOptions struct {
			Username string `json:"username"`
			Hostname string `json:"hostname"`
		} `json:"clientOptions"`
	}
	if err := c.readJSON(ctx, []string{cmdRepository, "status", cmdFlagJSON}, &status); err != nil {
		return "", hperrors.Wrap(err)
	}
	if info.Owner == status.ClientOptions.Username+"@"+status.ClientOptions.Hostname {
		return "", nil
	}

	if _, err := c.execCommand(ctx, []string{cmdMaintenance, cmdSet, "--owner=me"}); err != nil {
		return "", hperrors.Wrap(err)
	}
	return info.Owner, nil
}
