package fileservice

import (
	"github.com/localpaas/localpaas/localpaas_app/repository"
)

type FileService interface {
}

func NewFileService(
	settingRepo repository.SettingRepo,
) FileService {
	return &fileService{
		settingRepo: settingRepo,
	}
}

type fileService struct {
	settingRepo repository.SettingRepo
}
