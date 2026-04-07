package taskcronjobexec

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/services/ssl/letsencrypt"
)

func (e *Executor) sslGetLeClient(
	ssl *entity.SSLCert,
	data *sslRenewalTaskData,
) (*letsencrypt.Client, error) {
	data.Mu.Lock()
	defer data.Mu.Unlock()

	email := gofn.Coalesce(ssl.Email, config.Current.SSL.LeUserEmail)
	keyType := gofn.Coalesce(ssl.KeyType, base.SSLKeyTypeDefault)
	mapKey := fmt.Sprintf("email:%v:keysize:%v", email, keyType)

	if client := data.LeClients[mapKey]; client != nil {
		return client, nil
	}

	client, err := letsencrypt.NewClient(email, keyType, config.Current.DataPathSslLetsEncrypt())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data.LeClients[mapKey] = client

	return client, nil
}

func (e *Executor) sslGetNotification(
	ctx context.Context,
	db database.IDB,
	sslSetting *entity.Setting,
	eventIsSuccess bool,
	data *sslRenewalTaskData,
) (notifSetting *entity.Setting, err error) {
	ssl := sslSetting.MustAsSSLCert()
	if ssl.Notification == nil {
		return nil, nil
	}

	data.Mu.Lock()
	defer data.Mu.Unlock()

	notifID := gofn.If(eventIsSuccess, ssl.Notification.Success.ID, ssl.Notification.Failure.ID)
	if notifID != "" {
		notifSetting = data.RefObjects.RefSettings[notifID]
		if notifSetting != nil {
			return notifSetting, nil
		}
	} else if (eventIsSuccess && !ssl.Notification.SuccessUseDefault) ||
		(!eventIsSuccess && !ssl.Notification.FailureUseDefault) {
		return nil, nil
	}

	var scope *base.SettingScope
	switch {
	case sslSetting.BelongToApp != nil:
		scope = sslSetting.BelongToApp.GetSettingScope()
	case sslSetting.BelongToProject != nil:
		scope = sslSetting.BelongToProject.GetSettingScope()
	default:
		scope = base.NewSettingScopeGlobal()
	}

	if notifID != "" {
		notifSetting, err = e.settingRepo.GetByID(ctx, db, scope, base.SettingTypeNotification, notifID, true)
	} else {
		notifSetting, err = e.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
			bunex.SelectWhere("setting.is_default = TRUE"),
		)
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if notifSetting == nil {
		return nil, nil
	}

	// Load ref objects of the setting (otherwise we will have error of missing ref objects)
	refObjects, err := e.settingService.LoadReferenceObjects(ctx, db, scope, true,
		false, notifSetting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data.AddRefObjects(refObjects)
	data.RefObjects.RefSettings[notifSetting.ID] = notifSetting

	return notifSetting, nil
}
