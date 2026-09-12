package appuc

import (
	"context"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/logging"
)

// GetAppLogHistory reads an app's stored logs.
//
// The app is loaded by project and id from the path, which the handler has
// already authorized; that app - not anything in the request - is what the
// query is scoped to.
func (uc *UC) GetAppLogHistory(
	ctx context.Context,
	_ *basedto.Auth,
	req *appdto.GetAppLogHistoryReq,
) (*appdto.GetAppLogHistoryResp, error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if featureSettings.LoggingSettings != nil && !featureSettings.LoggingSettings.Enabled {
		return nil, hperrors.NewUnavailable("App logs")
	}

	req.ApplyDefaults(timeutil.NowUTC())
	resp, err := uc.loggingService.QueryAppLogs(ctx, uc.db, app, &loggingservice.AppLogQuery{
		Contains: req.Search,
		Levels:   req.LevelList(),
		Streams:  req.StreamList(),
		Start:    req.Start,
		End:      req.End,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appdto.GetAppLogHistoryResp{Data: toHistoryData(resp)}, nil
}

func toHistoryData(resp *logging.QueryResp) *appdto.AppLogHistoryDataResp {
	data := &appdto.AppLogHistoryDataResp{
		Logs:      make([]*tasklog.LogFrame, 0, len(resp.Entries)),
		Truncated: resp.Truncated,
	}
	for i := range resp.Entries {
		e := &resp.Entries[i]
		data.Logs = append(data.Logs, &tasklog.LogFrame{Type: historyFrameType(e), Data: e.Message, Ts: e.Time})
	}
	if resp.Truncated && len(resp.Entries) > 0 {
		// One nanosecond before the oldest line shown. VictoriaLogs' end is
		// inclusive, so the oldest line itself would come back again.
		next := resp.Entries[0].Time.Add(-time.Nanosecond)
		data.NextEnd = &next
	}
	return data
}

// historyFrameType colors a stored line the way the live viewer colors one: by
// the message's own level when it has one, by stream otherwise.
func historyFrameType(e *logging.LogEntry) tasklog.LogType {
	switch strings.ToLower(e.Level) {
	case "error", "fatal", "panic":
		return tasklog.LogTypeErr
	case "warn", "warning":
		return tasklog.LogTypeWarn
	case "debug", "trace":
		return tasklog.LogTypeDebug
	case "":
		if e.Stream == "stderr" {
			return tasklog.LogTypeErr
		}
	}
	return tasklog.LogTypeOut
}
