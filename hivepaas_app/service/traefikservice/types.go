package traefikservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type AppConfigData struct {
	HttpSettings *entity.AppHttpSettings
	RefObjects   *entity.RefObjects
}
