package hpappservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type AppReleaseInfo struct {
	Stable *ReleaseInfo `json:"stable"`
	Beta   *ReleaseInfo `json:"beta"`
}

type ReleaseInfo struct {
	base.ReleaseInfo

	// System specific flag
	CanUpdate bool `json:"canUpdate"`
}
