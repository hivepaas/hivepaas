package settingservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type CreateEvent struct {
	Setting *entity.Setting
}

func (s *settingService) OnCreate(
	ctx context.Context,
	db database.IDB,
	event *CreateEvent,
) (err error) {
	// Save SSL cert/key files in a directory for using later
	if event.Setting.Type == base.SettingTypeSSLCert {
		err = s.PersistSSLCertFiles(true, event.Setting)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
