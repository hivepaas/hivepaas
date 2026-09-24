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

// checkDockerAPI refuses a request whose apps would be given the Docker API,
// unless the caller may grant that. The gate is the one capabilities have, Write
// on the Cluster module: the app's children run on the cluster's nodes, as many
// as its limits allow, running what its images allow.
func (uc *UC) checkDockerAPI(ctx context.Context, auth *basedto.Auth, apps []*appToProvision) error {
	block := specmodel.SingletonBlockName(base.SettingTypeAppDockerAPI)
	var asking []string
	for _, target := range apps {
		if _, found := target.result.Doc.Settings[block]; found {
			asking = append(asking, target.name)
		}
	}
	if len(asking) == 0 {
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
	return hperrors.Wrap(hperrors.ErrUnauthorized).
		WithExtraDetail("%s: giving an app the Docker API needs Write permission on the Cluster module",
			strings.Join(asking, ", ")).
		WithMsgLog("creating an app from a template with the Docker API requires Write on the Cluster module")
}
