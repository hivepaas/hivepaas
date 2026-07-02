package traefikuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

type UC struct {
	traefikService traefikservice.Service
}

func New(
	traefikService traefikservice.Service,
) *UC {
	return &UC{
		traefikService: traefikService,
	}
}
