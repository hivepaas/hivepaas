package gitlab

import (
	"context"
	"net/http"

	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// CanWrite answers whether a user is a member of a project, directly or by its
// groups, at Developer or above - who may push to it. A user who is not a
// member may not.
func (c *Client) CanWrite(ctx context.Context, pid any, userID int64) (bool, error) {
	member, resp, err := c.client.ProjectMembers.GetInheritedProjectMember(pid, userID, gogitlab.WithContext(ctx))
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return member.AccessLevel >= gogitlab.DeveloperPermissions, nil
}
