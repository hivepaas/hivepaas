package clusterservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type PersistingClusterData struct {
	UpsertingSettings []*entity.Setting
}
