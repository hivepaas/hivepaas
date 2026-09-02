package sysstatusdto

import (
	"database/sql"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type GetDBStatsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *DBStatsResp  `json:"data"`
}

// DBStatsResp exposes the connection pool counters so pool sizing can be decided from measurement
// rather than guesswork.
type DBStatsResp struct {
	// MaxOpenConnections is the configured ceiling; OpenConnections and InUse are the live usage.
	MaxOpenConnections int `json:"maxOpenConnections"`
	OpenConnections    int `json:"openConnections"`
	InUse              int `json:"inUse"`
	Idle               int `json:"idle"`

	// WaitCount is the number of times a caller had to wait for a free connection, and
	// WaitDuration the total time spent waiting. These two answer whether the pool is a
	// bottleneck: a WaitCount that stays near zero means raising MaxOpenConns would change
	// nothing.
	WaitCount         int64  `json:"waitCount"`
	WaitDurationMs    int64  `json:"waitDurationMs"`
	WaitDurationHuman string `json:"waitDurationHuman"`

	// Connections closed by each limit, useful for spotting churn: a large MaxIdleClosed means
	// the idle pool is too small for the load.
	MaxIdleClosed     int64 `json:"maxIdleClosed"`
	MaxIdleTimeClosed int64 `json:"maxIdleTimeClosed"`
	MaxLifetimeClosed int64 `json:"maxLifetimeClosed"`
}

func TransformDBStats(stats sql.DBStats) *DBStatsResp {
	return &DBStatsResp{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,

		WaitCount:         stats.WaitCount,
		WaitDurationMs:    stats.WaitDuration.Milliseconds(),
		WaitDurationHuman: stats.WaitDuration.Round(time.Millisecond).String(),

		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}
