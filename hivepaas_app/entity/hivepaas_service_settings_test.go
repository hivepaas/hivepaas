package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func TestProxySettingsClientIPDepth(t *testing.T) {
	t.Run("a configured proxy", func(t *testing.T) {
		s := &HivePaaSProxySettings{ProxyProvider: "cloudflare", ProxyHops: 3}
		assert.Equal(t, 3, s.ClientIPDepth())
	})

	// With nothing in front, X-Forwarded-For is written by the caller. Reading a
	// position in it would let every caller name their own address.
	t.Run("no proxy means the peer address", func(t *testing.T) {
		s := &HivePaaSProxySettings{ProxyHops: 3}
		assert.Zero(t, s.ClientIPDepth())
	})

	t.Run("a proxy without hops is not guessed at", func(t *testing.T) {
		s := &HivePaaSProxySettings{ProxyProvider: "cloudflare"}
		assert.Zero(t, s.ClientIPDepth())
	})

	t.Run("a nil receiver", func(t *testing.T) {
		var s *HivePaaSProxySettings
		assert.Zero(t, s.ClientIPDepth())
		assert.False(t, s.HasProxy())
	})
}

func hivePaaSServiceSetting(t *testing.T, version int, data *HivePaaSService) *Setting {
	t.Helper()
	setting := &Setting{
		ID:      gofn.Must(ulid.NewStringULID()),
		Scope:   base.HivepaasScope,
		Type:    base.SettingTypeHivePaaSService,
		Version: version,
	}
	setting.MustSetData(data)
	return setting
}

func TestHivePaaSServiceMigrate(t *testing.T) {
	// Upgrading must not change how any existing deployment identifies its
	// callers, so the depth the label builder used to hard-code is written in.
	t.Run("carries the old hard-coded depth into a configured proxy", func(t *testing.T) {
		data := &HivePaaSService{ProxySettings: HivePaaSProxySettings{
			ProxyProvider: "cloudflare",
			TrustedIPs:    []string{"173.245.48.0/20"},
		}}
		setting := hivePaaSServiceSetting(t, 1, data)

		hasChange, err := data.Migrate(setting)
		assert.NoError(t, err)
		assert.True(t, hasChange)
		assert.Equal(t, LegacyProxyHops, data.ProxySettings.ProxyHops)
		assert.Equal(t, CurrentHivePaaSServiceVersion, setting.Version)
	})

	// Writing a depth into an install with nothing in front would tell it to read
	// a header the caller writes.
	t.Run("leaves an install with no proxy alone", func(t *testing.T) {
		data := &HivePaaSService{}
		setting := hivePaaSServiceSetting(t, 1, data)

		_, err := data.Migrate(setting)
		assert.NoError(t, err)
		assert.Zero(t, data.ProxySettings.ProxyHops)
	})

	t.Run("does not overwrite a depth already set", func(t *testing.T) {
		data := &HivePaaSService{ProxySettings: HivePaaSProxySettings{
			ProxyProvider: "cloudflare", ProxyHops: 5,
		}}
		setting := hivePaaSServiceSetting(t, 1, data)

		_, err := data.Migrate(setting)
		assert.NoError(t, err)
		assert.Equal(t, 5, data.ProxySettings.ProxyHops)
	})

	t.Run("is a no-op at the current version", func(t *testing.T) {
		data := &HivePaaSService{ProxySettings: HivePaaSProxySettings{ProxyProvider: "cloudflare"}}
		setting := hivePaaSServiceSetting(t, CurrentHivePaaSServiceVersion, data)

		hasChange, err := data.Migrate(setting)
		assert.NoError(t, err)
		assert.False(t, hasChange)
		assert.Zero(t, data.ProxySettings.ProxyHops)
	})
}
