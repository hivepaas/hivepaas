package logginguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
)

type UC struct {
	db             *database.DB
	settingRepo    repository.SettingRepo
	loggingService loggingservice.Service
}

// New builds the usecase. fx supplies the arguments from the provider list in
// registry/provides.go; *database.DB is what database.NewDB provides.
func New(
	db *database.DB,
	settingRepo repository.SettingRepo,
	loggingService loggingservice.Service,
) *UC {
	return &UC{db: db, settingRepo: settingRepo, loggingService: loggingService}
}
