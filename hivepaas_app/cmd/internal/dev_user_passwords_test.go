package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/userservice/userserviceimpl"
)

// usersInMemory is a user repository holding its users in a slice.
type usersInMemory struct {
	repository.UserRepo
	users   []*entity.User
	updated []string
}

func (r *usersInMemory) List(context.Context, database.IDB, *basedto.Paging, ...bunex.SelectQueryOption) (
	[]*entity.User, *basedto.PagingMeta, error) {
	return r.users, nil, nil
}

func (r *usersInMemory) Update(_ context.Context, _ database.IDB, user *entity.User,
	_ ...bunex.UpdateQueryOption) error {
	r.updated = append(r.updated, user.Username)
	return nil
}

func TestDevModeGivesEveryUserThePassword(t *testing.T) {
	const seedHash = "VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ="
	repo := &usersInMemory{users: []*entity.User{
		{ID: "u1", Username: "tiendc", Password: seedHash},
		{ID: "u2", Username: "member1", Password: seedHash},
	}}
	users := userserviceimpl.New(nil, nil, nil, nil, nil, repo, nil, nil)

	count, err := resetUserPasswords(context.Background(), nil, repo, users, "Dev-Shared#2026")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, []string{"tiendc", "member1"}, repo.updated)
	for _, user := range repo.users {
		assert.NotEqual(t, seedHash, user.Password)
		assert.NoError(t, users.VerifyPassword(user, "Dev-Shared#2026"), user.Username)
		assert.Error(t, users.VerifyPassword(user, "abc123"), "the seed's password no longer signs in")
	}
}

func TestDevModeRefusesAWeakSharedPassword(t *testing.T) {
	repo := &usersInMemory{users: []*entity.User{{ID: "u1", Username: "tiendc"}}}
	users := userserviceimpl.New(nil, nil, nil, nil, nil, repo, nil, nil)

	_, err := resetUserPasswords(context.Background(), nil, repo, users, "abc123")
	assert.Error(t, err)
	assert.Empty(t, repo.updated, "nothing is written")
}
