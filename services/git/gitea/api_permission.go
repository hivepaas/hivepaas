package gitea

import (
	"net/http"
	"slices"

	gogitea "code.gitea.io/sdk/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// CanWrite answers whether a user may write to a repository - its owner, an
// admin, or a collaborator or team member with write access. A user Gitea does
// not know of may not.
func (c *Client) CanWrite(owner, repo, user string) (bool, error) {
	result, resp, err := c.client.CollaboratorPermission(owner, repo, user)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return slices.Contains(writerAccessModes, result.Permission), nil
}

// writerAccessModes are the access modes that may write to a repository.
var writerAccessModes = []gogitea.AccessMode{gogitea.AccessModeWrite, gogitea.AccessModeAdmin,
	gogitea.AccessModeOwner}
