package apptemplateuc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
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
		if granted := grantedCapabilities(target); len(granted) > 0 {
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

// grantedCapabilities names, for a person to read, what one app's document asks
// the host for. It is empty for an app whose document carries no capabilities.
func grantedCapabilities(target *appToProvision) []string {
	doc := target.rendered.Result.Doc
	if doc == nil || doc.Deployment == nil || doc.Deployment.Resources == nil {
		return nil
	}
	capabilities := doc.Deployment.Resources.Capabilities
	if capabilities == nil {
		return nil
	}
	granted := make([]string, 0, len(capabilities.CapabilityAdd))
	granted = append(granted, capabilities.CapabilityAdd...)
	if capabilities.EnableGPU {
		granted = append(granted, docker.CapabilityGPU)
	}
	for _, described := range []struct {
		what  string
		count int
	}{
		{"sysctls", len(capabilities.Sysctls)}, {"ulimits", len(capabilities.Ulimits)},
	} {
		if described.count > 0 {
			granted = append(granted, described.what)
		}
	}
	if len(granted) == 0 && len(capabilities.CapabilityDrop) == 0 && capabilities.OomScoreAdj == 0 {
		return nil
	}
	if len(granted) == 0 {
		// A block that only drops capabilities or nudges the OOM score grants
		// nothing, but it is still the privileged block, and the gate is the
		// block's rather than each field's.
		granted = append(granted, string(specmodel.BlockDeploymentResources)+".capabilities")
	}
	return granted
}
