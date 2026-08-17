package commandpipeuc

import (
	"context"
	"errors"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/strutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

func (uc *UC) CreateCommandPipeFromTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *commandpipedto.CreateCommandPipeFromTemplateReq,
) (*commandpipedto.CreateCommandPipeFromTemplateResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			srcCmd, tgtCmd, pipeName, err := uc.createTemplatedCommands(ctx, db, req)
			if err != nil {
				return apperrors.Wrap(err)
			}

			if req.Name == "" {
				pipeName, err = uc.calcCommandPipeName(ctx, db, req.Scope, currentSettingType, pipeName)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}

			commandPipe := &entity.CommandPipe{
				SourceCommand: entity.ObjectID{ID: srcCmd.ID},
				TargetCommand: entity.ObjectID{ID: tgtCmd.ID},
			}
			pData.Setting.Name = gofn.Coalesce(req.Name, pipeName)
			err = pData.Setting.SetData(commandPipe)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &commandpipedto.CreateCommandPipeFromTemplateResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) createTemplatedCommands(
	ctx context.Context,
	db database.IDB,
	req *commandpipedto.CreateCommandPipeFromTemplateReq,
) (srcCmd, tgtCmd *entity.Setting, pipeName string, err error) {
	cmdKind := "pipe"
	var e1, e2 error

	switch req.CommandType {
	case "postgres", "postgresql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "pg_dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "pg_restore", cmdKind)
		pipeName = req.CommandType + " (pg_dump -> pg_restore)"

	case "mysql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "mysqldump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "mysql", cmdKind)
		pipeName = req.CommandType + " (mysqldump -> mysql)"

	case "mariadb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "mariadb-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "mariadb", cmdKind)
		pipeName = req.CommandType + " (mariadb-dump -> mariadb)"

	case "mongo", "mongodb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "mongodump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "mongorestore", cmdKind)
		pipeName = req.CommandType + " (mongodump -> mongorestore)"

	case "redis":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "redis-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "redis-restore", cmdKind)
		pipeName = req.CommandType + " (redis-dump -> redis-restore)"

	case "clickhouse":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "clickhouse-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "clickhouse-restore", cmdKind)
		pipeName = req.CommandType + " (clickhouse-dump -> clickhouse-restore)"

	case "sqlite", "sqlite3":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "sqlite-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "sqlite-restore", cmdKind)
		pipeName = req.CommandType + " (sqlite-dump -> sqlite-restore)"

	case "sqlserver", "mssql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "sqlserver-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "sqlserver-restore", cmdKind)
		pipeName = req.CommandType + " (sqlserver-dump -> sqlserver-restore)"

	case "influx", "influxdb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "influx-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "influx-restore", cmdKind)
		pipeName = req.CommandType + " (influx-dump -> influx-restore)"

	case "elasticsearch", "opensearch":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "elasticsearch-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "elasticsearch-restore", cmdKind)
		pipeName = req.CommandType + " (elasticsearch-dump -> elasticsearch-restore)"

	case "dolt":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "dolt-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "dolt-restore", cmdKind)
		pipeName = req.CommandType + " (dolt-dump -> dolt-restore)"

	case "neon", "neon-postgres":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, "neon-dump", cmdKind)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, "neon-restore", cmdKind)
		pipeName = req.CommandType + " (neon-dump -> neon-restore)"

	default:
		return nil, nil, "",
			apperrors.NewUnsupported(fmt.Sprintf("Argument '%s'", req.CommandType))
	}
	err = errors.Join(e1, e2)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err)
	}

	if req.Scope != nil {
		srcCmd.Scope = req.Scope.ScopeType
		srcCmd.ObjectID = req.Scope.ScopeObjectID()
		tgtCmd.Scope = req.Scope.ScopeType
		tgtCmd.ObjectID = req.Scope.ScopeObjectID()
	}
	srcCmd.AvailInProjects = req.AvailInProjects
	tgtCmd.AvailInProjects = req.AvailInProjects

	// Prevent repeated names by adding suffix
	srcCmd.Name, err = uc.calcCommandPipeName(ctx, db, req.Scope, base.SettingTypeCommandTemplate, srcCmd.Name)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err)
	}
	tgtCmd.Name, err = uc.calcCommandPipeName(ctx, db, req.Scope, base.SettingTypeCommandTemplate, tgtCmd.Name)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err)
	}

	err = uc.SettingRepo.UpsertMulti(ctx, db, []*entity.Setting{srcCmd, tgtCmd},
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err)
	}

	return srcCmd, tgtCmd, pipeName, nil
}

func (uc *UC) calcCommandPipeName(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	settingType base.SettingType,
	baseName string,
) (suggestedPipeName string, err error) {
	for i := 1; i <= 10; i++ {
		name := baseName
		if i > 1 {
			name = fmt.Sprintf("%s (%d)", baseName, i)
		}
		err = uc.checkNameConflict(ctx, db, scope, settingType, name)
		if err == nil {
			return name, nil
		}
	}
	return "", apperrors.Wrap(apperrors.ErrUnavailable).
		WithParam("Name", "Command pipe name space")
}

func (uc *UC) checkNameConflict(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	settingType base.SettingType,
	name string,
) (err error) {
	if name == "" {
		return nil
	}
	setting, err := uc.SettingRepo.GetByName(ctx, db, scope, settingType, name, false,
		bunex.SelectColumns("id", "name"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if setting != nil {
		return apperrors.NewAlreadyExist(strutil.ToPascalCase(string(settingType))).
			WithMsgLog("%s '%s' already exists", settingType, setting.Name)
	}
	return nil
}
