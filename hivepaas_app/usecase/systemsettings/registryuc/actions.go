package registryuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryuc/registrydto"
)

// ProbeRegistryDomain reports what answers at an address, so that the dashboard
// can warn about a proxy in front of the registry before a build meets its upload
// limit.
func (uc *UC) ProbeRegistryDomain(
	ctx context.Context,
	_ *basedto.Auth,
	req *registrydto.ProbeDomainReq,
) (*registrydto.ProbeDomainResp, error) {
	probe, err := uc.registryService.ProbeDomain(ctx, req.Domain)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registrydto.ProbeDomainResp{Data: &registrydto.DomainProbeResp{
		Reached:  probe.Reached,
		Proxied:  probe.Proxied,
		Evidence: probe.Evidence,
	}}, nil
}

// CheckRegistryPush uploads a large blob through the public domain and throws it
// away. It is the only answer about a body limit that is not a guess.
func (uc *UC) CheckRegistryPush(
	ctx context.Context,
	_ *basedto.Auth,
	req *registrydto.PushCheckReq,
) (*registrydto.PushCheckResp, error) {
	setting, err := uc.loadStoredSetting(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	result, err := uc.registryService.CheckPush(ctx, uc.DB, &registryservice.PushCheckReq{
		Setting: setting,
		Bytes:   req.Bytes,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registrydto.PushCheckResp{Data: &registrydto.PushCheckDataRes{
		OK:         result.OK,
		StatusCode: result.StatusCode,
		Detail:     result.Detail,
		ElapsedMs:  result.Elapsed.Milliseconds(),
	}}, nil
}

// RotateRegistryCredential issues a new password and leaves the previous one
// working until the grace period ends.
func (uc *UC) RotateRegistryCredential(
	ctx context.Context,
	_ *basedto.Auth,
) (*registrydto.RotateCredentialResp, error) {
	setting, err := uc.loadStoredSetting(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := uc.registryService.RotateCredential(ctx, uc.DB, &registryservice.RotateCredentialReq{
		Setting: setting,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registrydto.RotateCredentialResp{Data: &registrydto.RotateCredentialDataRes{
		GraceEndsAt: resp.GraceEndsAt,
	}}, nil
}

// loadStoredSetting reads the configuration these actions act on. Nothing is
// provisioned until it exists, so its absence is the same answer as "not
// configured".
func (uc *UC) loadStoredSetting(ctx context.Context) (*entity.Setting, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.DB, entity.NewObjectScopeGlobal(),
		currentSettingType, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}
