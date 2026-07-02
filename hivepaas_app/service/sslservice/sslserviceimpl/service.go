package sslserviceimpl

import "github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"

func New() sslservice.Service {
	return &service{}
}

type service struct {
}
