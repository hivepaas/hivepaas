package dbservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/repository"
)

type DBService interface {
	MigrateData(ctx context.Context, db database.IDB) error
}

func NewDBService(
	dataMigrationRepo repository.DataMigrationRepo,
	settingRepo repository.SettingRepo,
) DBService {
	return &dbService{
		dataMigrationRepo: dataMigrationRepo,
		settingRepo:       settingRepo,
	}
}

type dbService struct {
	dataMigrationRepo repository.DataMigrationRepo
	settingRepo       repository.SettingRepo
}
