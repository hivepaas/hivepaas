package sessionuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// fakeUserRepo answers GetByID with one user; nothing else of it is called.
type fakeUserRepo struct {
	repository.UserRepo
	user *entity.User
}

func (f *fakeUserRepo) GetByID(
	_ context.Context, _ database.IDB, _ string, _ ...bunex.SelectQueryOption,
) (*entity.User, error) {
	return f.user, nil
}

// The session says which release the installation runs and the API level it
// answers, and below which level a CLI may not write: the CLI reads it on login
// and before a command that needs what an older server lacks.
func TestGetMeSaysWhatTheServerRuns(t *testing.T) {
	user := &entity.User{ID: "user1", Email: "dev@example.com", Status: base.UserStatusActive}
	uc := &UC{userRepo: &fakeUserRepo{user: user}}

	resp, err := uc.GetMe(context.Background(), &basedto.User{User: user}, &sessiondto.GetMeReq{})

	assert.NoError(t, err)
	assert.Equal(t, &sessiondto.ServerInfoResp{
		Version:        systemappservice.CurrentRelease().AppVersion,
		APILevel:       base.APILevel,
		MinCLIAPILevel: base.APILevel,
	}, resp.Data.Server)
}
