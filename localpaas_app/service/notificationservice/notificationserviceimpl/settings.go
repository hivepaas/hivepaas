package notificationserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (s *service) GetNotificationForEvent(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	eventSetting *entity.BaseEventNotification,
	eventSuccess bool,
	refObjects *entity.RefObjects,
) (_ *entity.Notification, err error) {
	if eventSetting == nil {
		return nil, nil
	}
	if refObjects == nil {
		refObjects = entity.NewRefObjects()
	}

	notifSettingID := gofn.If(eventSuccess, eventSetting.Success.ID, eventSetting.Failure.ID)
	if notifSettingID == "" {
		if (eventSuccess && !eventSetting.SuccessUseDefault) || (!eventSuccess && !eventSetting.FailureUseDefault) {
			return nil, nil
		}
	}

	var setting *entity.Setting
	if notifSettingID == "" {
		setting, err = s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
			bunex.SelectWhere("setting.is_default = TRUE"),
		)
	} else {
		if loadedSetting, ok := refObjects.RefSettings[notifSettingID]; ok && loadedSetting != nil {
			setting = loadedSetting
		} else {
			setting, err = s.settingRepo.GetByID(ctx, db, scope, base.SettingTypeNotification,
				notifSettingID, true)
		}
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}

	// Load ref objects of the setting (otherwise we will have error of missing ref objects)
	refs, err := s.settingService.LoadReferenceObjects(ctx, db, scope, true,
		false, setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	refObjects.AddRefObjects(refs)
	refObjects.RefSettings[setting.ID] = setting

	return setting.MustAsNotification(), nil
}

func (s *service) GetDefaultNotification(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	refObjects *entity.RefObjects,
	errorIfRefObjectsUnavail bool,
) (*entity.Notification, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeNotification, true,
		bunex.SelectWhere("setting.is_default = TRUE"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil
	}

	if refObjects != nil {
		// Load ref objects of the setting (otherwise we will have error of missing ref objects)
		refs, err := s.settingService.LoadReferenceObjects(ctx, db, scope, true,
			errorIfRefObjectsUnavail, setting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		refObjects.AddRefObjects(refs)
		if refObjects.RefSettings == nil {
			refObjects.RefSettings = make(map[string]*entity.Setting)
		}
		refObjects.RefSettings[setting.ID] = setting
	}

	return setting.MustAsNotification(), nil
}

func (s *service) BuildTitlePrefix(
	project *entity.Project,
	app *entity.App,
	user *entity.User,
) string {
	switch {
	case app != nil:
		return fmt.Sprintf("[%s][%s]", project.Name, app.Name)
	case project != nil:
		return fmt.Sprintf("[%s]", project.Name)
	case user != nil:
		return fmt.Sprintf("[User][%s]", gofn.Coalesce(user.FullName, user.Username))
	default:
		return "[System]"
	}
}
