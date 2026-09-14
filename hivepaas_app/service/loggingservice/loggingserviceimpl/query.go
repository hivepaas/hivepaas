package loggingserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	cfg, err := s.loadEnabledSettings(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	ep, err := queryEndpoint(cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	backend, err := s.newBackend(logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{Endpoint: *ep}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := backend.Query(ctx, &logging.QueryReq{
		Match:   []logging.FieldMatch{{Field: vlagent.AttrField(appservice.LabelLogAppID), Value: app.ID}},
		Search:  q.Search,
		Levels:  q.Levels,
		Streams: q.Streams,
		Start:   q.Start,
		End:     q.End,
		Limit:   q.Limit,
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
	cfg, err := s.loadEnabledSettings(ctx, db)
	if errors.Is(err, hperrors.ErrLoggingNotEnabled) {
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
			excludedReason := isAppExcluded(app, &inspect.Service)
			if excludedReason != "" {
				return &loggingservice.AppHistory{Reason: historyReason(excludedReason)}, nil
			}
		}
	}
	return &loggingservice.AppHistory{Available: true, Retention: retentionOf(cfg)}, nil
}

// retentionOf is how long stored logs are kept, when HivePaaS is the one
// keeping them.
//
// An unmanaged backend expires on a schedule HivePaaS has no way to read, so it
// reports nothing there rather than a number the operator never set.
func retentionOf(cfg *entity.LoggingSettings) timeutil.Duration {
	if !cfg.Backend.Managed || cfg.Backend.VictoriaLogs == nil {
		return 0
	}
	return cfg.Backend.VictoriaLogs.Retention
}

func historyReason(r loggingservice.ExcludedReason) loggingservice.HistoryUnavailableReason {
	if r == loggingservice.ExcludedReasonDriverUnreadable {
		return loggingservice.HistoryReasonDriverUnreadable
	}
	return loggingservice.HistoryReasonIdentityMissing
}

// queryEndpoint is where to read from: the backend HivePaaS runs, reached by
// service name over the API's private network, or the one the operator named.
func queryEndpoint(cfg *entity.LoggingSettings) (*logging.Endpoint, error) {
	if cfg.Backend.Managed {
		return &logging.Endpoint{URL: backendBaseURL()}, nil
	}
	if !hasQueryEndpoint(cfg) {
		return nil, hperrors.Wrap(hperrors.ErrLoggingQueryEndpointMissing)
	}
	return toEndpoint(cfg.Backend.Query)
}

// hasQueryEndpoint says whether there is anywhere to read from.
func hasQueryEndpoint(cfg *entity.LoggingSettings) bool {
	return cfg.Backend.Managed || (cfg.Backend.Query != nil && cfg.Backend.Query.URL != "")
}
