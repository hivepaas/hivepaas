package approutingserviceimpl

import (
	"context"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

const (
	// routingApplyRetryMax is how many times an apply that swarm refused as out of
	// sequence is tried again, and routingApplyRetryDelay how long it waits before
	// each - the service's field of that name, which tests shorten.
	routingApplyRetryMax   = 3
	routingApplyRetryDelay = 500 * time.Millisecond

	// swarmUpdateOutOfSequence is what swarm answers an update carrying a version the
	// service has since moved on from (swarmkit's store.ErrSequenceConflict).
	swarmUpdateOutOfSequence = "update out of sequence"
)

type applyAppRoutingData struct {
	*approutingservice.ApplyAppRoutingReq
}

// ApplyRoutingSettings implements approutingservice.Service.
//
// It retries, because the service update at its end can be refused as "update out
// of sequence": swarm rejects an update carrying the version an inspect returned
// when something else wrote to the service in between. Right after a service is
// created or deployed, swarm's own allocator is still writing to it - about one
// create in three is refused on a developer machine, and so is a routing change
// saved right after a deployment - and any other write in between does the same.
//
// Each attempt starts again from a fresh inspect, also when the caller handed in
// the service. The spec is built by steps that write into it - the labels, the
// networks, the update terms - so the only way to a spec that carries both the
// current version and those changes is to run the steps again on what swarm holds
// now. That is also why the update is not done with ServiceUpdateFunc: it would
// re-inspect on its own and push the fresh spec without them.
//
// Running the steps again is safe. Each sets what the routing settings call for
// over whatever is there, rather than adding to it: the Traefik labels are
// replaced as a set and the routing network attached only when it is not already,
// the certificate and Traefik files are written over or removed, and Traefik's
// ports are opened only when they are not open yet. The refused update itself
// changed nothing. So a repeated attempt ends where a first one that went through
// would have.
//
// Only that refusal is retried. It is the one failure that says nothing is wrong
// with the apply itself; anything else would fail the same way again, and trying
// would only hold up the error - inside a transaction, for most callers.
func (s *service) ApplyRoutingSettings(
	ctx context.Context,
	db database.IDB,
	req *approutingservice.ApplyAppRoutingReq,
) (*approutingservice.ApplyAppRoutingResp, error) {
	var err error
	for attempt := range routingApplyRetryMax + 1 {
		// A copy, so that what one attempt loads and builds is never what the next
		// starts from, and the caller's request is left as it was given.
		attemptReq := *req
		if attempt > 0 {
			timer := time.NewTimer(s.routingApplyRetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, hperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
			// The service the last attempt built on carries the version swarm has
			// just refused, whether the caller handed it in or it was inspected.
			attemptReq.Service = nil
		}

		var resp *approutingservice.ApplyAppRoutingResp
		resp, err = s.applyRoutingSettingsOnce(ctx, db, &attemptReq)
		if err == nil {
			return resp, nil
		}
		if !isUpdateOutOfSequence(err) {
			return nil, hperrors.Wrap(err)
		}
	}
	return nil, hperrors.Wrap(err)
}

// applyRoutingSettingsOnce is a single attempt of ApplyRoutingSettings.
func (s *service) applyRoutingSettingsOnce(
	ctx context.Context,
	db database.IDB,
	req *approutingservice.ApplyAppRoutingReq,
) (*approutingservice.ApplyAppRoutingResp, error) {
	data := &applyAppRoutingData{
		ApplyAppRoutingReq: req,
	}

	err := s.loadAppRoutingData(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.applyRoutingSettings(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &approutingservice.ApplyAppRoutingResp{Service: data.Service}, nil
}

// isUpdateOutOfSequence tells whether err is swarm refusing an update for the
// version it carries.
//
// The text is all there is to go by. The daemon passes swarm's error on as a 500
// whose message is "rpc error: code = Unknown desc = update out of sequence", the
// docker client gives that no type of its own, and hperrors.NewInfra only adds the
// infra code beside it. Every wrapper on the way here keeps the message in its own,
// so matching the whole error's text finds it however deep it is.
func isUpdateOutOfSequence(err error) bool {
	return err != nil && strings.Contains(err.Error(), swarmUpdateOutOfSequence)
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
