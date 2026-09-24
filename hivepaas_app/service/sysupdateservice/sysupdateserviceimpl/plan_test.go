package sysupdateserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

func TestPlanComponent(t *testing.T) {
	blockDb := []string{base.HivepaasDbKey}
	tests := []struct {
		name     string
		key      string
		current  string
		deployed bool
		target   string
		block    []string
		change   sysupdateservice.Change
		backup   bool
		traffic  bool
	}{
		{
			name: "a newer image of the same major is an update", key: base.HivepaasCacheKey,
			current: "redis:8.6-alpine", deployed: true, target: "redis:8.8-alpine",
			change: sysupdateservice.ChangeUpdate,
		},
		{
			name: "the image already running is left alone", key: base.HivepaasCacheKey,
			current: "redis:8.6-alpine", deployed: true, target: "redis:8.6-alpine",
			change: sysupdateservice.ChangeNone,
		},
		{
			name: "an older image is not moved to", key: base.HivepaasTraefikKey,
			current: "traefik:v3.8", deployed: true, target: "traefik:v3.7",
			change: sysupdateservice.ChangeNone,
		},
		{
			name: "a postgres major is made from the backup", key: base.HivepaasDbKey,
			current: "postgres:18.3-alpine", deployed: true, target: "postgres:19.0-alpine",
			change: sysupdateservice.ChangeMajor, backup: true,
		},
		{
			name: "a major the release blocks is refused", key: base.HivepaasDbKey,
			current: "postgres:18.3-alpine", deployed: true, target: "postgres:19.0-alpine", block: blockDb,
			change: sysupdateservice.ChangeBlocked,
		},
		{
			name: "a blocked component's minor move is not blocked", key: base.HivepaasDbKey,
			current: "postgres:18.3-alpine", deployed: true, target: "postgres:18.4-alpine", block: blockDb,
			change: sysupdateservice.ChangeUpdate,
		},
		{
			name: "moving the proxy interrupts traffic", key: base.HivepaasTraefikKey,
			current: "traefik:v3.7", deployed: true, target: "traefik:v3.8",
			change: sysupdateservice.ChangeUpdate, traffic: true,
		},
		{
			name: "a component not running is not deployed", key: base.HivepaasRegistryKey,
			target: "ghcr.io/project-zot/zot:v2.1.22",
			change: sysupdateservice.ChangeNotDeployed,
		},
		{
			name: "a release naming no image changes nothing", key: base.HivepaasRegistryKey,
			current: "ghcr.io/project-zot/zot:v2.1.21", deployed: true,
			change: sysupdateservice.ChangeNone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := planComponent(tt.key, tt.current, tt.deployed, tt.target, tt.block)

			assert.Equal(t, tt.change, got.Change)
			assert.Equal(t, tt.backup, got.RequiresBackup)
			assert.Equal(t, tt.traffic, got.InterruptsTraffic)
			assert.NotEmpty(t, got.Reason)
		})
	}
}
