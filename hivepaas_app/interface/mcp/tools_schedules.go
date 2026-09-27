package mcp

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// scheduleDescs describe a schedule, as the sched job endpoints take one;
// prefix is the path of the schedule in the request.
func scheduleDescs(prefix string) map[string]string {
	return map[string]string{
		prefix + "cronExpr": "a cron expression - minute, hour, day of month, month, day of week, such as " +
			"0 3 * * 1-5 - or a descriptor such as @daily, @hourly or @every 90m; read in the time zone of " +
			"initialTime. Give it or interval, not both",
		prefix + "interval": "instead of cronExpr: a duration such as 90m, 12h or 1d; the runs are initialTime, " +
			"then every interval after it",
		prefix + "initialTime": "when the schedule starts, RFC 3339 - its offset is the zone cronExpr is read " +
			"in, such as 2026-09-27T00:00:00+07:00; a fixed offset, so no daylight saving. Now, in UTC, " +
			"when not given; not more than a year ago",
	}
}

// ---- explain_schedule ----

func explainScheduleTool() Tool {
	descs := scheduleDescs("")
	descs["count"] = "how many runs to answer, 1-10"
	return readToolWith("explain_schedule", "Explain a schedule",
		"POST /settings/sched-jobs/calc-next-runs. The next runs of a cron expression or an interval from "+
			"initialTime, as HivePaaS computes them for a scheduled job: data is their times, in the offset "+
			"of initialTime. Use it to check a schedule before setting it up.",
		bodyInput(underNothing, &schedjobdto.CalcNextRunsReq{}, descs, []string{"count"},
			"endTime"), // taken, but not applied to the runs answered
		func(ctx context.Context, call *Call, in map[string]any) (*schedjobdto.CalcNextRunsResp, error) {
			body := &schedjobdto.CalcNextRunsReq{}
			if err := decodeBody(in, body); err != nil {
				return nil, err
			}
			var resp schedjobdto.CalcNextRunsResp
			if err := call.Post(ctx, "/settings/sched-jobs/calc-next-runs", body, &resp); err != nil {
				return nil, err
			}
			return &resp, nil
		})
}
