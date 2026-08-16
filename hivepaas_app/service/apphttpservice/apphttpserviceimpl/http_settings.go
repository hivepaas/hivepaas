package apphttpserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apphttpservice"
)

type applyAppHttpData struct {
	*apphttpservice.ApplyAppHttpReq
}

func (s *service) ApplyHttpSettings(
	ctx context.Context,
	db database.IDB,
	req *apphttpservice.ApplyAppHttpReq,
) (resp *apphttpservice.ApplyAppHttpResp, err error) {
	resp = &apphttpservice.ApplyAppHttpResp{}
	data := &applyAppHttpData{
		ApplyAppHttpReq: req,
	}

	err = s.loadAppHttpData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applyHttpSettings(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Service = data.Service
	return resp, nil
}

func (s *service) loadAppHttpData(
	ctx context.Context,
	db database.IDB,
	data *applyAppHttpData,
) (err error) {
	// Load reference objects
	refObjectIDs := data.HttpSettings.GetRefObjectIDs()

	err = s.settingService.LoadRefObjectsByIDs(ctx, db, &data.RefObjects, data.App.GetObjectScope(),
		true, refObjectIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.Service == nil {
		inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
		if err != nil {
			return apperrors.Wrap(err)
		}
		data.Service = &inspect.Service
	}

	return nil
}
