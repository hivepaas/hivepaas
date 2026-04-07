package taskcronjobexec

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
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/notificationservice"
	"github.com/localpaas/localpaas/services/ssl/letsencrypt"
)

const (
	sslHandlingBatchSize       = 100
	sslHandlingConcurrentTasks = 5
)

type sslRenewalTaskData struct {
	*taskData
	TaskOutput *entity.TaskSSLRenewalOutput
	LeClients  map[string]*letsencrypt.Client
	Mu         *sync.Mutex
}

type sslRenewalTaskItem struct {
	Setting              *entity.Setting
	Renewal              bool
	ExpiringNotifyOnly   bool
	RenewalError         error
	SettingSavedToDB     bool
	ExpiringNotifMsgData *notificationservice.BaseMsgDataSSLExpiringNotification
	RenewalNotifMsgData  *notificationservice.BaseMsgDataSSLRenewalNotification
}

//nolint:gocognit
func (e *Executor) cronExecSSLRenew(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) error {
	renewConfig := data.RefObjects.RefSettings[data.CronJob.TargetSetting.ID]
	if renewConfig == nil {
		return apperrors.NewNotFound("SSL renew settings")
	}

	taskData := &sslRenewalTaskData{
		taskData:   data,
		TaskOutput: &entity.TaskSSLRenewalOutput{},
		LeClients:  make(map[string]*letsencrypt.Client),
		Mu:         &sync.Mutex{},
	}
	timeNow := timeutil.NowUTC()

	// TODO (high): renew self-signed SSL cert if needs to

	taskArgs := gofn.Coalesce(gofn.Must(data.Task.ArgsAsSSLRenewal()), &entity.TaskSSLRenewalArgs{})
	offset, limit := 0, sslHandlingBatchSize
	for {
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
		if len(taskArgs.TargetSSLs) > 0 {
			listOpts = append(listOpts,
				bunex.SelectWhereIn("setting.id IN (?)", taskArgs.TargetSSLs.ToIDStringSlice()...))
		} else {
			listOpts = append(listOpts,
				bunex.SelectOffset(offset), bunex.SelectLimit(limit))
		}

		sslSettings, _, err := e.settingRepo.List(ctx, db, nil, nil, listOpts...)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if len(sslSettings) == 0 {
			break
		}
		offset += limit

		taskItems := make([]*sslRenewalTaskItem, 0, len(sslSettings))
		for _, setting := range sslSettings {
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
			taskItems = append(taskItems, &sslRenewalTaskItem{
				Setting: setting,
			})
		}

		_ = gofn.ExecTaskFuncEx(ctx, sslHandlingConcurrentTasks, false,
			func(ctx context.Context, taskItem *sslRenewalTaskItem) error {
				ssl := taskItem.Setting.MustAsSSLCert()
				switch {
				case e.sslShouldNotifyOfExpiration(ssl, timeNow):
					taskItem.ExpiringNotifyOnly = true
				case e.sslShouldRenew(ssl, timeNow):
					taskItem.Renewal = true
					taskItem.RenewalError = e.sslRenew(ctx, taskItem.Setting, taskData)
					if taskItem.RenewalError != nil {
						return apperrors.Wrap(taskItem.RenewalError)
					}
				}
				return nil
			},
			taskItems...)

		// NOTE: Ignore the error of the current processing batch to continue with remaining SSLs
		_ = e.sslSaveUpdatedSettings(ctx, taskItems, timeNow, taskData)

		// Send notifications for the result
		_ = e.sslNotifyOfResult(ctx, db, taskItems, taskData)

		if len(taskArgs.TargetSSLs) > 0 {
			break
		}
	}

	// Assign back the result output
	data.Task.MustSetOutput(taskData.TaskOutput)

	// Reload traefik config
	if len(taskData.TaskOutput.RenewedSSLs) > 0 {
		_ = e.traefikService.ReloadTraefikConfig(ctx, true)
	}

	return nil
}

func (e *Executor) sslShouldRenew(
	ssl *entity.SSLCert,
	timeNow time.Time,
) bool {
	return ssl.AutoRenew &&
		(!ssl.RenewableFrom.IsZero() && timeNow.After(ssl.RenewableFrom) && timeNow.Before(ssl.ExpireAt) ||
			ssl.RenewableFrom.IsZero())
}

func (e *Executor) sslRenew(
	ctx context.Context,
	setting *entity.Setting,
	data *sslRenewalTaskData,
) (err error) {
	ssl := setting.MustAsSSLCert()

	startTime := timeutil.NowUTC()
	defer func() {
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewWarnFrame(fmt.Sprintf(
				"Obtaining certificate from %v for SSL %v failed with error: %v",
				ssl.CertType, setting.ID, err.Error()), applog.TsNow))
		} else {
			duration := timeutil.NowUTC().Sub(startTime)
			_ = data.LogStore.Add(ctx, applog.NewOutFrame(fmt.Sprintf(
				"Obtaining certificate from %v for SSL %v finished in %v",
				ssl.CertType, setting.ID, duration), applog.TsNow))
		}
	}()

	if ssl.CertType == base.SSLCertTypeLetsEncrypt {
		return e.sslRenewByLetsEncrypt(ctx, ssl, data)
	}

	return apperrors.NewUnsupported(fmt.Sprintf("SSL type '%v'", ssl.CertType))
}

func (e *Executor) sslShouldNotifyOfExpiration(
	ssl *entity.SSLCert,
	timeNow time.Time,
) bool {
	return !ssl.AutoRenew && !ssl.NotifyFrom.IsZero() &&
		timeNow.After(ssl.NotifyFrom) && timeNow.Before(ssl.ExpireAt)
}

func (e *Executor) sslSaveUpdatedSettings(
	ctx context.Context,
	taskItems []*sslRenewalTaskItem,
	timeNow time.Time,
	data *sslRenewalTaskData,
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
	err = transaction.Execute(ctx, e.db, func(db database.Tx) error {
		// Reloads SSL settings to see if we should update them with the renewed cert
		reloadedSettings, err := e.settingRepo.ListByIDs(ctx, db, nil, settingIDs, true,
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
				_ = data.LogStore.Add(ctx, applog.NewWarnFrame(fmt.Sprintf(
					"Skip renewing SSL %v as of concurrent modification", setting.ID), applog.TsNow))
				continue
			}
			setting.UpdatedAt = timeNow
			setting.UpdateVer++
			persistingSettings = append(persistingSettings, setting)
			continue
		}

		err = e.settingRepo.UpsertMulti(ctx, db, persistingSettings,
			entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = e.settingService.PersistSSLCertFiles(true, persistingSettings...)
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
func (e *Executor) sslNotifyOfResult(
	ctx context.Context,
	db database.IDB,
	taskItems []*sslRenewalTaskItem,
	data *sslRenewalTaskData,
) (err error) {
	_ = gofn.ExecTaskFuncEx(ctx, sslHandlingConcurrentTasks, false,
		func(ctx context.Context, item *sslRenewalTaskItem) error {
			if item.ExpiringNotifyOnly {
				err := e.sslNotifyOfExpiration(ctx, db, item, data)
				if err != nil {
					_ = data.LogStore.Add(ctx, applog.NewWarnFrame(fmt.Sprintf(
						"Notifying of expiring SSL %v failed with error: %v",
						item.Setting.ID, err.Error()), applog.TsNow))
					return apperrors.Wrap(err)
				}
				return nil
			}
			if item.Renewal {
				err := e.sslNotifyOfRenewal(ctx, db, item, data)
				if err != nil {
					_ = data.LogStore.Add(ctx, applog.NewWarnFrame(fmt.Sprintf(
						"Notifying of renewed SSL %v failed with error: %v",
						item.Setting.ID, err.Error()), applog.TsNow))
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
