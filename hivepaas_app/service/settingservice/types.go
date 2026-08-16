package settingservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type PersistingSettingData struct {
	UpsertingSettings []*entity.Setting
	UpsertingAccesses []*entity.ACLPermission
	DeletingAccesses  []*entity.ACLPermission
}
