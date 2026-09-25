package settingmountuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

// ListSettingMountSources is the parts registry, and whether the caller may
// mount the sensitive parts: what the screen needs to offer an entry.
func (uc *UC) ListSettingMountSources(
	ctx context.Context, auth *basedto.Auth, _ *settingmountdto.ListSettingMountSourcesReq,
) (*settingmountdto.ListSettingMountSourcesResp, error) {
	mayMount, err := uc.PermissionManager.MayRevealSecrets(ctx, uc.DB, auth)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &settingmountdto.ListSettingMountSourcesResp{MayMountSensitive: mayMount}
	for _, typ := range settingmountservice.SourceTypes() {
		source := &settingmountdto.SettingMountSource{Type: typ}
		for _, part := range settingmountservice.PartsOf(typ) {
			source.Parts = append(source.Parts, &settingmountdto.SettingMountPart{
				Name: part.Name, Required: part.Required, Secret: part.Secret, Gated: part.Gated})
		}
		resp.Data = append(resp.Data, source)
	}
	return resp, nil
}
