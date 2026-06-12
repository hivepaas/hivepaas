package sslrenewalserviceimpl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/localpaas_app/service/sslrenewalservice"
	"github.com/localpaas/localpaas/services/ssl/acme"
)

const (
	sslHandlingBatchSize       = 100
	sslHandlingConcurrentTasks = 5
)

type sslRenewalData struct {
	*sslrenewalservice.SSLRenewalReq
	TaskOutput  *entity.TaskSSLRenewalOutput
	AcmeClients map[string]*acme.Client
	Mu          *sync.Mutex
}

type sslRenewalDataItem struct {
	Setting              *entity.Setting
	Renewal              bool
	ExpiringNotifyOnly   bool
	RenewalError         error
	SettingSavedToDB     bool
	ExpiringNotifMsgData *notificationservice.TemplateDataSSLExpiring
	RenewalNotifMsgData  *notificationservice.TemplateDataSSLRenewal
}

func (s *service) SSLRenew(
	ctx context.Context,
	db database.Tx,
	req *sslrenewalservice.SSLRenewalReq,
) (resp *sslrenewalservice.SSLRenewalResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	resp = &sslrenewalservice.SSLRenewalResp{
		SkipResultNotification: true,
	}
	data := &sslRenewalData{
		SSLRenewalReq: req,
		TaskOutput:    &entity.TaskSSLRenewalOutput{},
		AcmeClients:   make(map[string]*acme.Client),
		Mu:            &sync.Mutex{},
	}

	// Load all SSL providers in the system
	err = s.loadSSLProviders(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	renewalArgs := gofn.Coalesce(gofn.Must(data.Task.ArgsAsSSLRenewal()), &entity.TaskSSLRenewalArgs{})
	timeNow := timeutil.NowUTC()
	offset, limit := 0, sslHandlingBatchSize
	for {
		taskItems, err := s.loadSSLCerts(ctx, db, renewalArgs, offset, limit, timeNow)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if len(taskItems) == 0 {
			break
		}
		offset += limit

		_ = gofn.ExecTaskFuncEx(ctx, sslHandlingConcurrentTasks, false,
			func(ctx context.Context, taskItem *sslRenewalDataItem) error {
				ssl := taskItem.Setting.MustAsSSLCert()
				switch {
				case s.sslShouldNotifyOfExpiration(ssl, timeNow):
					taskItem.ExpiringNotifyOnly = true
				case s.sslShouldRenew(ssl, timeNow):
					taskItem.Renewal = true
					taskItem.RenewalError = s.sslRenew(ctx, taskItem.Setting, data)
					if taskItem.RenewalError != nil {
						return apperrors.Wrap(taskItem.RenewalError)
					}
				}
				return nil
			},
			taskItems...)

		// NOTE: Ignore the error of the current processing batch to continue with remaining SSLs
		_ = s.sslSaveUpdatedSettings(ctx, taskItems, timeNow, data)

		// Send notifications for the result
		_ = s.sslNotifyOfResult(ctx, db, taskItems, data)

		if len(renewalArgs.TargetSSLs) > 0 {
			break
		}
	}

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	// Reload traefik config
	if len(data.TaskOutput.RenewedSSLs) > 0 {
		_ = s.traefikService.ReloadTraefikConfig(ctx, true)
	}

	return resp, nil
}

func (s *service) loadSSLProviders(
	ctx context.Context,
	db database.IDB,
	data *sslRenewalData,
) (err error) {
	providerSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLProvider),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, setting := range providerSettings {
		data.RefObjects.RefSettings[setting.ID] = setting
	}

	return nil
}

func (s *service) loadSSLCerts(
	ctx context.Context,
	db database.IDB,
	renewalArgs *entity.TaskSSLRenewalArgs,
	offset, limit int,
	timeNow time.Time,
) (_ []*sslRenewalDataItem, err error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereGroup(
			bunex.SelectWhereGroup(
				bunex.SelectWhere("(setting.data->'autoRenew')::BOOL = TRUE"),
				bunex.SelectWhereGroup(
					bunex.SelectWhere("setting.data->>'renewableFrom' IS NULL"),
					bunex.SelectWhereOr("(setting.data->>'renewableFrom')::TIMESTAMPTZ < ?", timeNow),
				),
			),
			bunex.SelectWhereOrGroup(
				bunex.SelectWhere("(setting.data->'autoRenew')::BOOL != TRUE"),
				bunex.SelectWhere("(setting.data->>'notifyFrom')::TIMESTAMPTZ < ?", timeNow),
			),
		),
		bunex.SelectRelation("BelongToProject",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp",
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		),
		bunex.SelectRelation("BelongToApp.Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	}
	if len(renewalArgs.TargetSSLs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhereIn("setting.id IN (?)", renewalArgs.TargetSSLs.ToIDStringSlice()...))
	} else {
		listOpts = append(listOpts, bunex.SelectOffset(offset), bunex.SelectLimit(limit))
	}

	sslCertSettings, _, err := s.settingRepo.List(ctx, db, nil, nil, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(sslCertSettings) == 0 {
		return nil, nil
	}

	taskItems := make([]*sslRenewalDataItem, 0, len(sslCertSettings))
	for _, setting := range sslCertSettings {
		if setting.BelongToApp != nil {
			setting.BelongToProject = setting.BelongToApp.Project
		}
		project := setting.BelongToProject
		app := setting.BelongToApp
		if app != nil && app.Status != base.AppStatusActive {
			continue
		}
		if project != nil && project.Status != base.ProjectStatusActive {
			continue
		}
		taskItems = append(taskItems, &sslRenewalDataItem{
			Setting: setting,
		})
	}
	return taskItems, nil
}

func (s *service) sslShouldRenew(
	ssl *entity.SSLCert,
	timeNow time.Time,
) bool {
	return ssl.AutoRenew &&
		(!ssl.RenewableFrom.IsZero() && timeNow.After(ssl.RenewableFrom) && timeNow.Before(ssl.ExpireAt) ||
			ssl.RenewableFrom.IsZero())
}

func (s *service) sslRenew(
	ctx context.Context,
	setting *entity.Setting,
	data *sslRenewalData,
) (err error) {
	ssl := setting.MustAsSSLCert()

	startTime := timeutil.NowUTC()
	defer func() {
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf(
				"Obtaining certificate from %v for SSL %v failed with error: %v",
				ssl.CertType, setting.ID, err.Error()), tasklog.TsNow))
		} else {
			duration := timeutil.NowUTC().Sub(startTime)
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf(
				"Obtaining certificate from %v for SSL %v finished in %v",
				ssl.CertType, setting.ID, duration), tasklog.TsNow))
		}
	}()

	switch ssl.CertType { //nolint:exhaustive
	case base.SSLCertTypeLetsEncrypt, base.SSLCertTypeZeroSSL, base.SSLCertTypeGoogleTS:
		err = s.sslRenewByACME(ctx, ssl, data)
	case base.SSLCertTypeSelfSigned:
		err = s.sslRenewSelfSignedCert(ctx, ssl, data)
	default:
		return apperrors.NewUnsupported(apperrors.Fmt("SSL type '%v'", ssl.CertType))
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	setting.MustSetData(ssl)
	return nil
}

func (s *service) sslShouldNotifyOfExpiration(
	ssl *entity.SSLCert,
	timeNow time.Time,
) bool {
	return !ssl.AutoRenew && !ssl.NotifyFrom.IsZero() &&
		timeNow.After(ssl.NotifyFrom) && timeNow.Before(ssl.ExpireAt)
}

func (s *service) sslSaveUpdatedSettings(
	ctx context.Context,
	taskItems []*sslRenewalDataItem,
	timeNow time.Time,
	data *sslRenewalData,
) (err error) {
	sslSettings := make([]*entity.Setting, 0, len(taskItems))
	for _, taskItem := range taskItems {
		if taskItem.Renewal && taskItem.RenewalError == nil {
			sslSettings = append(sslSettings, taskItem.Setting)
		}
	}
	settingIDs := entityutil.ExtractIDs(sslSettings)
	if len(settingIDs) == 0 {
		return nil
	}
	var persistingSettings []*entity.Setting
	// Open a new transaction to save updated settings
	err = transaction.Execute(ctx, s.db, func(db database.Tx) error {
		// Reloads SSL settings to see if we should update them with the renewed cert
		reloadedSettings, err := s.settingRepo.ListByIDs(ctx, db, nil, settingIDs, true,
			bunex.SelectFor("UPDATE"),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}

		reloadedSettingMap := entityutil.SliceToIDMap(reloadedSettings)
		persistingSettings = make([]*entity.Setting, 0, len(reloadedSettings))
		for _, setting := range sslSettings {
			reloadedSetting := reloadedSettingMap[setting.ID]
			if reloadedSetting == nil || reloadedSetting.UpdateVer != setting.UpdateVer {
				_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf(
					"Skip renewing SSL %v as of concurrent modification", setting.ID), tasklog.TsNow))
				continue
			}
			setting.UpdatedAt = timeNow
			setting.UpdateVer++
			persistingSettings = append(persistingSettings, setting)
			continue
		}

		err = s.settingRepo.UpsertMulti(ctx, db, persistingSettings,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.sslService.WriteCertFiles(true, persistingSettings...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, sslSetting := range persistingSettings {
		data.TaskOutput.RenewedSSLs = append(data.TaskOutput.RenewedSSLs,
			&entity.ObjectID{ID: sslSetting.ID})
	}
	return nil
}

//nolint:unparam
func (s *service) sslNotifyOfResult(
	ctx context.Context,
	db database.IDB,
	taskItems []*sslRenewalDataItem,
	data *sslRenewalData,
) (err error) {
	_ = gofn.ExecTaskFuncEx(ctx, sslHandlingConcurrentTasks, false,
		func(ctx context.Context, item *sslRenewalDataItem) error {
			if item.ExpiringNotifyOnly {
				err := s.sslNotifyForExpiration(ctx, db, item, data)
				if err != nil {
					_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf(
						"Notification of expiring SSL %v failed with error: %v",
						item.Setting.ID, err.Error()), tasklog.TsNow))
					return apperrors.Wrap(err)
				}
				return nil
			}
			if item.Renewal {
				err := s.sslNotifyForRenewal(ctx, db, item, data)
				if err != nil {
					_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf(
						"Notification of renewed SSL %v failed with error: %v",
						item.Setting.ID, err.Error()), tasklog.TsNow))
					return apperrors.Wrap(err)
				}
				return nil
			}
			return nil
		},
		taskItems...)

	for _, item := range taskItems {
		if item.ExpiringNotifyOnly {
			data.TaskOutput.ExpiringNotifiedSSLs = append(data.TaskOutput.ExpiringNotifiedSSLs,
				&entity.ObjectID{ID: item.Setting.ID})
		}
	}

	return nil
}
