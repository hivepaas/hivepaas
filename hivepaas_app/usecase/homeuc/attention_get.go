package homeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/homeuc/homedto"
)

// GetHomeAttention is what needs attention, narrowed to what the user may see.
//
// An item is shown to whoever may open the screen it leads to, and nobody else:
// an app's to whoever may read its env, a node's to whoever may read the cluster
// screens. Nothing is said of what is left out - not even how much - since that
// alone would tell a user that something is wrong where they cannot look.
func (uc *UC) GetHomeAttention(
	ctx context.Context,
	auth *basedto.Auth,
) (*homedto.GetHomeAttentionResp, error) {
	items, err := uc.attentionService.Items(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	visibility := uc.permissionManager.NewVisibility(uc.db, auth)
	data := &homedto.HomeAttentionDataResp{Items: make([]*homedto.AttentionItemResp, 0, len(items))}
	for _, item := range items {
		canRead, err := allows(ctx, visibility, item.Scope, base.ActionTypeRead)
		if err != nil {
			return nil, err
		}
		if !canRead {
			continue
		}
		canAct, err := allows(ctx, visibility, item.Scope, base.ActionTypeWrite)
		if err != nil {
			return nil, err
		}
		data.Items = append(data.Items, homedto.TransformAttentionItem(item, canAct))
	}
	return &homedto.GetHomeAttentionResp{Data: data}, nil
}

// allows asks what the screen an item leads to would ask.
func allows(
	ctx context.Context,
	visibility permission.Visibility,
	scope attentionservice.Scope,
	action base.ActionType,
) (bool, error) {
	var allowed bool
	var err error
	switch scope.Type {
	case attentionservice.ScopeApp:
		allowed, err = visibility.AllowsProjectEnv(ctx, scope.ProjectID, scope.ProjectEnv, action)
	case attentionservice.ScopeSystem:
		allowed, err = visibility.AllowsModule(ctx, base.ResourceModuleSystem, action)
	case attentionservice.ScopeCluster:
		allowed, err = visibility.AllowsModule(ctx, base.ResourceModuleCluster, action)
	default:
		return false, nil // a scope nobody decided on is shown to nobody
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return allowed, nil
}
