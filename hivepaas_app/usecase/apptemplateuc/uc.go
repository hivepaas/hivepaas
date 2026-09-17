package apptemplateuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
)

type UC struct {
	appTemplateService apptemplateservice.Service
}

func New(
	appTemplateService apptemplateservice.Service,
) *UC {
	return &UC{
		appTemplateService: appTemplateService,
	}
}
