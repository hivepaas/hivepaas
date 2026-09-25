package appsettingsuc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) ExecuteAppClone(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.ExecuteAppCloneReq,
) (*appsettingsdto.ExecuteAppCloneResp, error) {
	// The gate goes first, outside the transaction: its answer is recorded
	// whatever becomes of the clone.
	entries, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
		bunex.SelectWhere("setting.object_id = ?", req.AppID),
		bunex.SelectWhere("setting.inheritable = TRUE"),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dropGated, leftOut, err := uc.cloneMountsGate(ctx, auth, &entity.App{ID: req.AppID}, entries)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var data *executeAppCloneData
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &executeAppCloneData{DropGatedMounts: dropGated}
		err := uc.loadAppCloneSettingsForExecute(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		persistingData.UpsertingTasks = append(persistingData.UpsertingTasks, data.AppCloneTask)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if data != nil && data.AppCloneTask != nil {
		if err = uc.taskQueue.ScheduleTask(ctx, data.AppCloneTask); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	resp := &appsettingsdto.ExecuteAppCloneResp{}
	if len(leftOut) > 0 {
		resp.Meta = &basedto.Meta{Warning: "The clone is made without these setting mounts, whose files " +
			"you may not reveal: " + strings.Join(leftOut, ", ")}
	}
	return resp, nil
}

type executeAppCloneData struct {
	App          *entity.App
	AppCloneTask *entity.Task
	// DropGatedMounts is the gate's answer: the clone leaves out setting mounts
	// with a gated part.
	DropGatedMounts bool
}

func (uc *UC) loadAppCloneSettingsForExecute(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.ExecuteAppCloneReq,
	data *executeAppCloneData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppClone),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app
	cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)

	if cloneSetting == nil {
		return hperrors.NewNotFound("App clone settings")
	}
	cloneSettings := cloneSetting.MustAsAppCloneSettings()

	// Validate target app's name to be available
	appKey := projecthelper.CalcAppKey(cloneSettings.TargetName)
	targetEnv := gofn.Coalesce(cloneSettings.TargetEnv, app.ProjectEnv.Key)
	appGlobalKey := projecthelper.CalcAppGlobalKey(app.Project.Key, appKey, targetEnv)
	// App keys must be unique globally
	conflictApp, err := uc.appRepo.GetByGlobalKey(ctx, db, "", appGlobalKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if conflictApp != nil {
		return hperrors.NewAlreadyExist("App").
			WithMsgLog("app unique key '%s' already exists", appGlobalKey)
	}

	// Create a task for cloning the app
	cloneTask, err := uc.appCloneService.CreateAppCloneTask(data.App, data.DropGatedMounts)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.AppCloneTask = cloneTask

	return nil
}

// cloneMountsGate passes §7's gate for the gated parts of the app's inheritable
// entries, which a clone would copy: the person asking for the clone gets an app
// that reads them. A denial is an answer, not an error - the clone goes on
// without those entries, and says so.
func (uc *UC) cloneMountsGate(
	ctx context.Context, auth *basedto.Auth, app *entity.App, entries []*entity.Setting,
) (drop bool, leftOut []string, err error) {
	var grants []settingmountservice.Grant
	for _, setting := range entries {
		mount, parseErr := setting.AsAppSettingMount()
		if parseErr != nil {
			return false, nil, hperrors.Wrap(parseErr)
		}
		if g := settingmountservice.Grants(mount); len(g) > 0 {
			grants = append(grants, g...)
			leftOut = append(leftOut, setting.Name)
		}
	}
	if len(grants) == 0 {
		return false, nil, nil
	}
	detail, err := json.Marshal(map[string]any{"grants": grants})
	if err != nil {
		return false, nil, hperrors.Wrap(err)
	}
	err = uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIAction,
		ResType:  base.ResourceTypeSettingMount,
		ResID:    app.ID,
		ResName:  "clone (setting mounts)",
		Detail:   string(detail),
	})
	switch {
	case err == nil:
		return false, nil, nil
	case errors.Is(err, hperrors.ErrRevealSecretsDisabled),
		errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets):
		return true, leftOut, nil
	}
	return false, nil, hperrors.Wrap(err)
}
