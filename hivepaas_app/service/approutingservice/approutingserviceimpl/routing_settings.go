package approutingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

type applyAppRoutingData struct {
	*approutingservice.ApplyAppRoutingReq
}

func (s *service) ApplyRoutingSettings(
	ctx context.Context,
	db database.IDB,
	req *approutingservice.ApplyAppRoutingReq,
) (resp *approutingservice.ApplyAppRoutingResp, err error) {
	resp = &approutingservice.ApplyAppRoutingResp{}
	data := &applyAppRoutingData{
		ApplyAppRoutingReq: req,
	}

	err = s.loadAppRoutingData(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.applyRoutingSettings(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Service = data.Service
	return resp, nil
}

func (s *service) loadAppRoutingData(
	ctx context.Context,
	db database.IDB,
	data *applyAppRoutingData,
) (err error) {
	// Load reference objects
	refObjectIDs := data.RoutingSettings.GetRefObjectIDs()

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
