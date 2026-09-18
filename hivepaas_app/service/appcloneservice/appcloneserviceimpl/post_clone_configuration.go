package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
)

// applyClonedConfiguration applies to docker what the copy's settings say.
//
// It keeps what was created in docker, because that is what cleanupOnFail has to
// remove: a rolled-back clone leaves no record of the secrets and configs it
// made, and they would otherwise sit there belonging to nothing.
func (s *service) applyClonedConfiguration(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	resp, err := s.appProvisionService.ApplyAppConfiguration(ctx, db,
		&appprovisionservice.ApplyAppConfigurationReq{
			App:        data.DestApp,
			RefObjects: data.RefObjects,
		})
	if resp != nil {
		data.DestConfig, data.DestSecrets = resp.Configs, resp.Secrets
	}
	return hperrors.Wrap(err)
}
