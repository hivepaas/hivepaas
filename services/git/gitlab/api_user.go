package gitlab

import (
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (c *Client) GetCurrentUser() (*gogitlab.User, error) {
	if c.currentUser != nil {
		return c.currentUser, nil
	}
	user, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return nil, apperrors.New(err)
	}
	c.currentUser = user
	return user, nil
}
