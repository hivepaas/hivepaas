package appsettingsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func exposedDatabaseRouting(t *testing.T) *entity.Setting {
	t.Helper()
	routing := &entity.AppRoutingSettings{
		Port:           5432,
		ExposePublicly: true,
		Domains: []*entity.AppDomain{
			// No certificate yet, as while one is being obtained.
			{Enabled: true, Domain: "db.example.com", Protocol: base.NetworkProtocolTCP, ContainerPort: 5432},
			{Enabled: true, Domain: "db2.example.com", Protocol: base.NetworkProtocolTCP, ContainerPort: 5432,
				TLSPassthrough: true, SSLCert: entity.ObjectID{ID: "cert-1"}},
		},
	}
	setting := &entity.Setting{ID: "routing-1", Type: base.SettingTypeAppRouting, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(routing))
	return setting
}

func kindReq(port uint) *appsettingsdto.UpdateAppKindSettingsReq {
	return &appsettingsdto.UpdateAppKindSettingsReq{
		AppKindSettingsReq: &appsettingsdto.AppKindSettingsReq{
			Category: base.AppCategoryDatabase,
			Engine:   "postgres",
			Port:     port,
			Database: &appsettingsdto.AppKindDatabaseReq{DbName: "app", Username: "app"},
		},
	}
}

// Saving a database's kind settings - a new password, say - leaves its domains
// as the routing settings have them: the kind settings used to wipe them all
// when no certificate was picked, and turn TLS passthrough off.
func TestKindSettingsUpdateLeavesRoutingAlone(t *testing.T) {
	uc := &UC{}
	data := &updateAppKindSettingsData{App: &entity.App{ID: "app-1"}, RoutingSetting: exposedDatabaseRouting(t)}
	persisting := &persistingAppData{}

	uc.updateRoutingPortOnKindChange(kindReq(5432), data, persisting)

	assert.False(t, data.RoutingChanged)
	assert.Empty(t, persisting.UpsertingSettings)
	routing := data.RoutingSetting.MustAsAppRoutingSettings()
	assert.True(t, routing.ExposePublicly)
	assert.Len(t, routing.Domains, 2)
}

func TestKindSettingsUpdateMovesThePortOnly(t *testing.T) {
	uc := &UC{}
	data := &updateAppKindSettingsData{App: &entity.App{ID: "app-1"}, RoutingSetting: exposedDatabaseRouting(t)}
	persisting := &persistingAppData{}

	uc.updateRoutingPortOnKindChange(kindReq(6432), data, persisting)

	assert.True(t, data.RoutingChanged)
	assert.Len(t, persisting.UpsertingSettings, 1)
	routing := data.RoutingSetting.MustAsAppRoutingSettings()
	assert.Equal(t, 6432, routing.Port)
	assert.True(t, routing.ExposePublicly)
	assert.Len(t, routing.Domains, 2)
	for _, domain := range routing.Domains {
		assert.Equal(t, 6432, domain.ContainerPort)
	}
	assert.Empty(t, routing.Domains[0].SSLCert.ID)
	assert.Equal(t, "cert-1", routing.Domains[1].SSLCert.ID)
	assert.True(t, routing.Domains[1].TLSPassthrough)
}

func TestKindSettingsUpdateCreatesRoutingForANewPort(t *testing.T) {
	uc := &UC{}
	data := &updateAppKindSettingsData{App: &entity.App{ID: "app-1"}}
	persisting := &persistingAppData{}

	uc.updateRoutingPortOnKindChange(kindReq(5432), data, persisting)

	assert.True(t, data.RoutingChanged)
	routing := data.RoutingSetting.MustAsAppRoutingSettings()
	assert.Equal(t, 5432, routing.Port)
	assert.False(t, routing.ExposePublicly)
	assert.Empty(t, routing.Domains)
}
