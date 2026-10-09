package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

// settingsByID answers the settings it holds, by id, whatever the scope asked.
type settingsByID struct {
	repository.SettingRepo
	settings map[string]*entity.Setting
}

func (r *settingsByID) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if setting, ok := r.settings[id]; ok {
		return setting, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrSettingNotFound)
}

// linksTo has every setting linked to by one setting of an app.
type linksTo struct {
	repository.ResLinkRepo
}

func (linksTo) List(context.Context, database.IDB, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.ResLink, *basedto.PagingMeta, error) {
	return []*entity.ResLink{{SrcType: base.ResourceTypeApp, SrcID: "app_1"}}, nil, nil
}

func usagesUC() *BaseUC {
	return &BaseUC{
		SettingRepo: &settingsByID{settings: map[string]*entity.Setting{
			"conf_1": {ID: "conf_1", Type: base.SettingTypeConfigFile, ObjectID: "app_1"},
		}},
		ResLinkRepo: linksTo{},
	}
}

// An app's setting has its usages listed through the app's path.
func TestAnAppsSettingListsItsUsages(t *testing.T) {
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_1"}

	usages, err := usagesUC().GetSettingUsages(context.Background(), scope, "conf_1")

	assert.NoError(t, err)
	assert.Len(t, usages, 1)
}

// The path's checks are of the project or the app in it: a setting of another
// one is not found there, whoever may read that path.
func TestASettingOfAnotherAppIsNotFoundThroughThisOnesPath(t *testing.T) {
	scope := &entity.ObjectScope{ScopeType: base.ObjectScopeApp, AppID: "app_2"}

	usages, err := usagesUC().GetSettingUsages(context.Background(), scope, "conf_1")

	assert.ErrorIs(t, err, hperrors.ErrSettingNotFound)
	assert.Empty(t, usages)
}
