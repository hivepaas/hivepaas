package healthcheckserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/healthcheckservice"
)

type service struct {
}

func New() healthcheckservice.Service {
	return &service{}
}
