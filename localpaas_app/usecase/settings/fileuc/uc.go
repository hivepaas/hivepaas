package fileuc

import (
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type FileUC struct {
	*settings.BaseSettingUC
}

func NewFileUC(
	baseSettingUC *settings.BaseSettingUC,
) *FileUC {
	return &FileUC{
		BaseSettingUC: baseSettingUC,
	}
}
