package approutingserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return nil, hperrors.Wrap(err)
	}

	err = s.applyRoutingSettings(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
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

	// NOTE: this runs even when the caller already loaded the references, and it
	// re-queries exactly the ones the caller could not find - so a caller that
	// wants missing references tolerated has to say so here. Loading them
	// leniently before the call does nothing on its own.
	loadRefObjects := s.settingService.LoadRefObjectsByIDs
	if data.SkipMissingRefObjects {
		loadRefObjects = s.settingService.LoadRefObjectsByIDsSkipMissing
	}
	err = loadRefObjects(ctx, db, &data.RefObjects, data.App.GetObjectScope(), true, refObjectIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if data.Service == nil {
		inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.Service = &inspect.Service
	}

	return nil
}
