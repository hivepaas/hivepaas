package userservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	SkipCheckingCurrentPassword = ""
)

type PersistingUserData struct {
	UpsertingUsers      []*entity.User
	UpsertingSettings   []*entity.Setting
	UpsertingBinObjects []*entity.BinObject
	UpsertingAccesses   []*entity.ACLPermission
	DeletingAccesses    []*base.PermissionResource
}
