package loggingserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// QueryAppLogs searches one app's stored logs.
//
// The scope comes from app, which the caller loaded and authorized, and is the
// only Match the backend receives. Nothing in q is a filter on identity.
func (s *service) QueryAppLogs(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	q *loggingservice.AppLogQuery,
) (*logging.QueryResp, error) {
	if app == nil || app.ID == "" {
		// An empty value would match every line that carries no app at all.
		return nil, hperrors.Wrap(logging.ErrQueryScopeRequired)
	}
	cfg, err := s.loadEnabled(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ep, err := queryEndpoint(cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	backend, err := s.newBackend(logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{Endpoint: ep}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := backend.Query(ctx, &logging.QueryReq{
		Match:    []logging.FieldMatch{{Field: vlagent.AttrField(appservice.LabelLogAppID), Value: app.ID}},
		Contains: q.Contains,
		Levels:   q.Levels,
		Streams:  q.Streams,
		Start:    q.Start,
		End:      q.End,
		Limit:    q.Limit,
	})
	return resp, hperrors.Wrap(err)
}

// AppHistory says whether stored logs can be shown for the app.
//
// A service that no longer exists is not a reason: its past output is what
// history is for.
func (s *service) AppHistory(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (*loggingservice.AppHistory, error) {
	cfg, err := s.loadEnabled(ctx, db)
	if errors.Is(err, loggingservice.ErrNotEnabled) {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonDisabled}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Sources.Apps {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonAppsNotCollected}, nil
	}
	if !hasQueryEndpoint(cfg) {
		return &loggingservice.AppHistory{Reason: loggingservice.HistoryReasonNoQueryEndpoint}, nil
	}

	if app.ServiceID != "" {
		inspect, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if err == nil && inspect != nil {
			excluded := excludedApps([]*entity.App{app}, []swarm.Service{inspect.Service})
			if len(excluded) > 0 {
				return &loggingservice.AppHistory{Reason: historyReason(excluded[0].Reason)}, nil
			}
		}
	}
	return &loggingservice.AppHistory{Available: true}, nil
}

func historyReason(r loggingservice.ExcludedReason) loggingservice.HistoryUnavailableReason {
	if r == loggingservice.ExcludedReasonDriverUnreadable {
		return loggingservice.HistoryReasonDriverUnreadable
	}
	return loggingservice.HistoryReasonIdentityMissing
}

// loadEnabled returns the decrypted configuration, or ErrNotEnabled.
func (s *service) loadEnabled(ctx context.Context, db database.IDB) (*entity.Logging, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, hperrors.Wrap(loggingservice.ErrNotEnabled)
	}
	cfg, err := setting.AsLogging()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, hperrors.Wrap(loggingservice.ErrNotEnabled)
	}
	if err := cfg.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return cfg, nil
}

// queryEndpoint is where to read from: the backend HivePaaS runs, reached by
// service name over the API's private network, or the one the operator named.
func queryEndpoint(cfg *entity.Logging) (logging.Endpoint, error) {
	if cfg.Backend.Managed {
		return logging.Endpoint{URL: backendBaseURL()}, nil
	}
	if !hasQueryEndpoint(cfg) {
		return logging.Endpoint{}, hperrors.Wrap(loggingservice.ErrQueryEndpointMissing)
	}
	return toEndpoint(cfg.Backend.Query)
}

// hasQueryEndpoint says whether there is anywhere to read from.
func hasQueryEndpoint(cfg *entity.Logging) bool {
	return cfg.Backend.Managed || (cfg.Backend.Query != nil && cfg.Backend.Query.URL != "")
}
