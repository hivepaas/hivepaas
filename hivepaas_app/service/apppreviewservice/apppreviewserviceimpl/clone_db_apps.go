package apppreviewserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type cloneDBAppData struct {
	NewAppName string
	NewAppKey  string
	CloneResp  *appcloneservice.AppCloneResp
}

func (s *service) cloneDBApps(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) (err error) {
	if len(data.CloneDBApps) == 0 {
		return nil
	}

	cloneDBAppsData := make(map[string]*cloneDBAppData, len(data.CloneDBApps))
	for _, srcDBApp := range data.CloneDBApps {
		cloneDBAppsData[srcDBApp.ID] = &cloneDBAppData{}
	}
	data.CloneDBAppsData = cloneDBAppsData

	defer func() {
		if rev := recover(); rev != nil {
			err = errors.Join(err, apperrors.ErrPanic)
		}
		for _, cloneData := range cloneDBAppsData {
			if cloneData.CloneResp != nil && cloneData.CloneResp.OnCleanup != nil {
				_ = cloneData.CloneResp.OnCleanup(err)
			}
		}
	}()

	refReplacerParams := make([]string, 0)

	// Calculate new names/keys for the clone DB apps
	for _, srcDBApp := range data.CloneDBApps {
		newAppName := srcDBApp.Name
		if data.PullNumber > 0 {
			newAppName += fmt.Sprintf("-pr-%v", data.PullNumber)
		}
		newAppName += "-" + data.RandSuffix
		newAppKey := projecthelper.CalcAppKey(newAppName)

		appData := cloneDBAppsData[srcDBApp.ID]
		appData.NewAppName = newAppName
		appData.NewAppKey = newAppKey

		refReplacerParams = append(refReplacerParams, fmt.Sprintf("${%v.", srcDBApp.Key),
			fmt.Sprintf("${%v.", newAppKey))
	}
	data.CloneDBAppsEnvRefReplacer = strings.NewReplacer(refReplacerParams...)

	orderedDBApps, err := s.calcCloneOrderOfDBApps(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, srcDBApp := range orderedDBApps {
		err = s.cloneDBApp(ctx, db, srcDBApp, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) cloneDBApp(
	ctx context.Context,
	db database.IDB,
	dbApp *entity.App,
	data *createPreviewData,
) (err error) {
	currentApp := data.App

	var cloneResp *appcloneservice.AppCloneResp
	defer func() {
		if rev := recover(); rev != nil {
			err = errors.Join(err, apperrors.ErrPanic)
		}
		if cloneResp != nil && cloneResp.OnCleanup != nil { // Run the cleanup function
			_ = cloneResp.OnCleanup(err)
			cloneResp.OnCleanup = nil
		}
	}()

	// Create a task to clone the app
	cloneTask := &entity.Task{
		ID:       "fake-task-id",
		Scope:    base.ObjectScopeApp,
		ObjectID: currentApp.ID,
		Type:     base.TaskTypeAppClone,
	}
	cloneTask.MustSetArgs(&entity.TaskAppCloneArgs{SrcApp: entity.ObjectID{ID: dbApp.ID}})

	cloneResp, err = s.appCloneService.CloneApp(ctx, db, &appcloneservice.AppCloneReq{
		TaskExecData: &queue.TaskExecData{
			Task:       cloneTask,
			RefObjects: data.RefObjects,
			LogStore:   data.LogStore,
		},
		SrcApp: dbApp,
		OnCloneApp: func(targetApp, srcApp *entity.App) error {
			return s.onCloneDBApp(targetApp, srcApp, data)
		},
		OnCloneSetting: func(targetApp, srcApp *entity.App, setting *entity.Setting) (*entity.Setting, error) {
			return s.onCloneDBAppSetting(setting, data)
		},
		OnCloneService: func(targetApp, srcApp *entity.App, targetSvc, srcSvc *swarm.Service) error {
			return s.onCloneDBAppService(targetSvc, srcSvc)
		},
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Save the cloned app response
	if cloneResp != nil {
		data.CloneDBAppsData[dbApp.ID].CloneResp = cloneResp
	}

	return nil
}

//nolint:gocognit
func (s *service) calcCloneOrderOfDBApps(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) (_ []*entity.App, err error) {
	if len(data.CloneDBApps) <= 1 {
		return data.CloneDBApps, nil
	}

	dbAppIDs := entityutil.ExtractIDs(data.CloneDBApps)
	envSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhereIn("setting.object_id IN (?)", dbAppIDs...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Priority levels:
	// Score 0: No reference to any other DB app in data.CloneDBApps.
	// Score 1: References another DB app, but DOES NOT reference another_app.HIVEPAAS_HOST.
	// Score 2: References another_app.HIVEPAAS_HOST of another DB app.
	appPriorityMap := make(map[string]int, len(data.CloneDBApps))
	for _, app := range data.CloneDBApps {
		appPriorityMap[app.ID] = 0
	}

	refHostPart := "." + base.AppSystemEnvVarHost + "}"

	for _, envSetting := range envSettings {
		hasAnyRef := false
		hasHostRef := false
		currAppID := envSetting.ObjectID
		envVars := envSetting.MustAsEnvVars()
		for _, env := range envVars.Data {
			if env.IsLiteral || !s.envVarService.HasRef(env.Value) {
				continue
			}

			// TODO: if the current app references the original app of the `preview app`,
			// that should be error as we will clone the DB apps before creating the preview app.

			for _, otherApp := range data.CloneDBApps {
				if otherApp.ID == currAppID {
					continue
				}
				otherKey := otherApp.Key
				otherData := data.CloneDBAppsData[otherApp.ID]

				isRef := strings.Contains(env.Value, "${"+otherKey+".")
				if !isRef && otherData != nil && otherData.NewAppKey != "" {
					isRef = strings.Contains(env.Value, "${"+otherData.NewAppKey+".")
				}

				if isRef {
					hasAnyRef = true
					isHostRef := strings.Contains(env.Value, "${"+otherKey+refHostPart)
					if !isHostRef && otherData != nil && otherData.NewAppKey != "" {
						isHostRef = strings.Contains(env.Value, "${"+otherData.NewAppKey+refHostPart)
					}
					if isHostRef {
						hasHostRef = true
					}
				}
			}
		}

		score := 0
		if hasAnyRef {
			if hasHostRef {
				score = 2
			} else {
				score = 1
			}
		}
		appPriorityMap[currAppID] = score
	}

	orderedApps := make([]*entity.App, len(data.CloneDBApps))
	copy(orderedApps, data.CloneDBApps)

	sort.SliceStable(orderedApps, func(i, j int) bool {
		return appPriorityMap[orderedApps[i].ID] < appPriorityMap[orderedApps[j].ID]
	})

	return orderedApps, nil
}
