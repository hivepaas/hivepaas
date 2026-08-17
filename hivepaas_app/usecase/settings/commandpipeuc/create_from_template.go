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
	const cmdPipe = ".pipe"
	var e1, e2 error

	switch req.CommandKind {
	case "postgres", "postgresql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "pg_dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "pg_restore"+cmdPipe)
		pipeName = req.CommandKind + " (pg_dump -> pg_restore)"

	case "mysql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "mysqldump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "mysql"+cmdPipe)
		pipeName = req.CommandKind + " (mysqldump -> mysql)"

	case "mariadb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "mariadb-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "mariadb"+cmdPipe)
		pipeName = req.CommandKind + " (mariadb-dump -> mariadb)"

	case "mongo", "mongodb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "mongodump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "mongorestore"+cmdPipe)
		pipeName = req.CommandKind + " (mongodump -> mongorestore)"

	case "redis":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "redis-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "redis-restore"+cmdPipe)
		pipeName = req.CommandKind + " (redis-dump -> redis-restore)"

	case "clickhouse":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "clickhouse-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "clickhouse-restore"+cmdPipe)
		pipeName = req.CommandKind + " (clickhouse-dump -> clickhouse-restore)"

	case "sqlite", "sqlite3":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "sqlite-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "sqlite-restore"+cmdPipe)
		pipeName = req.CommandKind + " (sqlite-dump -> sqlite-restore)"

	case "sqlserver", "mssql":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "sqlserver-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "sqlserver-restore"+cmdPipe)
		pipeName = req.CommandKind + " (sqlserver-dump -> sqlserver-restore)"

	case "influx", "influxdb":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "influx-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "influx-restore"+cmdPipe)
		pipeName = req.CommandKind + " (influx-dump -> influx-restore)"

	case "elasticsearch", "opensearch":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "elasticsearch-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "elasticsearch-restore"+cmdPipe)
		pipeName = req.CommandKind + " (elasticsearch-dump -> elasticsearch-restore)"

	case "dolt":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "dolt-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "dolt-restore"+cmdPipe)
		pipeName = req.CommandKind + " (dolt-dump -> dolt-restore)"

	case "neon", "neon-postgres":
		srcCmd, e1 = uc.commandService.GetCommand(ctx, req.CommandType, "neon-dump"+cmdPipe)
		tgtCmd, e2 = uc.commandService.GetCommand(ctx, req.CommandType, "neon-restore"+cmdPipe)
		pipeName = req.CommandKind + " (neon-dump -> neon-restore)"

	default:
		return nil, nil, "",
			apperrors.NewUnsupported(fmt.Sprintf("Argument '%s'", req.CommandKind))
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
