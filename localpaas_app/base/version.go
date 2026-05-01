package base

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	VersionCodeV1 = "v000001" // Date: 2026-01-01
)

const CurrentVersion = VersionCodeV1

const StableVersionCode = VersionCodeV1

// TODO: update these info later
var StableVersion = &ReleaseInfo{
	ReleaseDate:  timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:   "v0.1.0",
	AppImage:     "localpaas/localpaas-dev:0.1.0",
	RedisImage:   "redis:8.6-alpine",
	DbImage:      "postgres:18.3-alpine",
	TraefikImage: "traefik:v3.6",
}

const BetaVersionCode = VersionCodeV1

// TODO: update these info later
var BetaVersion = &ReleaseInfo{
	ReleaseDate:  timeutil.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	AppVersion:   "v0.1.0-beta1",
	AppImage:     "localpaas/localpaas-dev:0.1.0",
	RedisImage:   "redis:8.6-alpine",
	DbImage:      "postgres:18.3-alpine",
	TraefikImage: "traefik:v3.6",
}

type ReleaseInfo struct {
	ReleaseDate  timeutil.Date `json:"releaseDate"`
	AppVersion   string        `json:"appVersion"`
	AppImage     string        `json:"appImage"`
	RedisImage   string        `json:"redisImage"`
	DbImage      string        `json:"dbImage"`
	TraefikImage string        `json:"traefikImage"`
}
