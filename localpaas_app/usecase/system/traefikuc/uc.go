package traefikuc

import (
	"github.com/localpaas/localpaas/localpaas_app/service/traefikservice"
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
