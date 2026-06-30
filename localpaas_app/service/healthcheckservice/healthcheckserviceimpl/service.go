package healthcheckserviceimpl

import (
	"github.com/localpaas/localpaas/localpaas_app/service/healthcheckservice"
)

type service struct {
}

func New() healthcheckservice.Service {
	return &service{}
}
