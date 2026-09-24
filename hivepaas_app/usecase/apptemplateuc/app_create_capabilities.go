package apptemplateuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// checkCapabilities refuses a request whose apps would be given more of the host
// than a container ordinarily gets, unless the caller may grant that.
//
// The gate is the one the app's resource settings screen applies to the same
// block: Write on the cluster module. Provisioning a template is how an app
// first gets capabilities, and it would otherwise be the way around a permission
// that stops anybody adding them afterwards.
//
// Dependencies count. They are created by the same request and run on the same
// nodes, and whoever cannot grant NET_ADMIN to an app cannot grant it to the
// database created alongside either.
func (uc *UC) checkCapabilities(
	ctx context.Context,
	auth *basedto.Auth,
	apps []*appToProvision,
) error {
	asked := map[string][]string{}
	for _, target := range apps {
		if granted := specmodel.GrantedCapabilities(target.result.Doc); len(granted) > 0 {
			asked[target.name] = granted
		}
	}
	if len(asked) == 0 {
		return nil
	}

	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleCluster,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if hasPerm {
		return nil
	}
	described := make([]string, 0, len(asked))
	for _, target := range apps {
		if granted, found := asked[target.name]; found {
			described = append(described, target.name+" asks for "+strings.Join(granted, ", "))
		}
	}
	return hperrors.Wrap(hperrors.ErrUnauthorized).
		WithExtraDetail("%s: granting this needs Write permission on the Cluster module",
			strings.Join(described, "; ")).
		WithMsgLog("creating an app from a template with capabilities requires Write on the Cluster module")
}
