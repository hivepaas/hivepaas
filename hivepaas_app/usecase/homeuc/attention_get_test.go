package homeuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
)

type fakeAttention struct{ items []*attentionservice.Item }

func (f *fakeAttention) Items(context.Context, database.IDB) ([]*attentionservice.Item, error) {
	return f.items, nil
}

// fakeVisibility grants what it is told: envs by "project/env/action", modules
// by "module/action".
type fakeVisibility struct{ grants map[string]bool }

func (f *fakeVisibility) AllowsModule(
	_ context.Context, module base.ResourceModule, action base.ActionType,
) (bool, error) {
	return f.grants[string(module)+"/"+string(action)], nil
}

func (f *fakeVisibility) AllowsProjectEnv(
	_ context.Context, projectID, env string, action base.ActionType,
) (bool, error) {
	return f.grants[projectID+"/"+env+"/"+string(action)], nil
}

type fakeManager struct {
	permission.Manager
	visibility *fakeVisibility
}

func (f *fakeManager) NewVisibility(database.IDB, *basedto.Auth) permission.Visibility {
	return f.visibility
}

func appItem(projectID, env, subject string) *attentionservice.Item {
	return &attentionservice.Item{
		Kind: attentionservice.KindAppNotRunning,
		Scope: attentionservice.Scope{
			Type: attentionservice.ScopeApp, ProjectID: projectID, ProjectEnv: env, AppID: subject,
		},
		Subject: subject,
	}
}

func subjects(t *testing.T, uc *UC) map[string]bool {
	t.Helper()
	resp, err := uc.GetHomeAttention(context.Background(), &basedto.Auth{})
	assert.NoError(t, err)
	shown := map[string]bool{}
	for _, item := range resp.Data.Items {
		shown[item.Subject] = item.CanAct
	}
	return shown
}

// Each item is shown to whoever may open the screen it leads to, and says
// whether they may act there.
func TestHomeAttentionShowsEachUserTheirPart(t *testing.T) {
	uc := &UC{
		attentionService: &fakeAttention{items: []*attentionservice.Item{
			appItem("prj_1", "dev", "a1"),
			appItem("prj_1", "prod", "a2"),
			appItem("prj_2", "dev", "a3"),
			{Kind: attentionservice.KindAppNotRunning, Scope: attentionservice.Scope{Type: attentionservice.ScopeSystem},
				Subject: "registry"},
			{Kind: attentionservice.KindNodeDown, Scope: attentionservice.Scope{Type: attentionservice.ScopeCluster},
				Subject: "appstack-1"},
			{Kind: attentionservice.KindNodeDown, Subject: "no scope"},
		}},
		permissionManager: &fakeManager{visibility: &fakeVisibility{grants: map[string]bool{
			"prj_1/dev/read":    true,
			"prj_1/dev/write":   true,
			"prj_2/dev/read":    true,
			"mod::cluster/read": true,
		}}},
	}

	shown := subjects(t, uc)

	assert.Equal(t, map[string]bool{"a1": true, "a3": false, "appstack-1": false}, shown)
}

// Nothing to show is an empty list, not a missing one.
func TestHomeAttentionIsAnEmptyListWhenAllIsWell(t *testing.T) {
	uc := &UC{
		attentionService:  &fakeAttention{},
		permissionManager: &fakeManager{visibility: &fakeVisibility{}},
	}

	resp, err := uc.GetHomeAttention(context.Background(), &basedto.Auth{})

	assert.NoError(t, err)
	assert.NotNil(t, resp.Data.Items)
	assert.Empty(t, resp.Data.Items)
}
