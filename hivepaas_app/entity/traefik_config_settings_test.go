package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// The whole probation flow goes through MustSetData and a later parse, and both
// panic rather than return on a type the registry does not know. A parser that
// was never registered, or registered under the wrong type, is therefore a crash
// on the first traefik config change - not a compile error and not a test failure
// anywhere else.
func TestTraefikConfigSettingRoundTrip(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeTraefikConfig, Version: CurrentTraefikConfigVersion}
	setting.MustSetData(&TraefikConfig{Args: []string{"traefik", "--accesslog=true"}})

	snapshot := SettingSnapshotOf(setting)

	setting.MustSetData(&TraefikConfig{Args: []string{"traefik"}})
	changed, err := setting.AsTraefikConfig()
	assert.NoError(t, err)
	assert.Equal(t, []string{"traefik"}, changed.Args)

	snapshot.RestoreTo(setting)
	restored, err := setting.AsTraefikConfig()
	assert.NoError(t, err)
	assert.Equal(t, []string{"traefik", "--accesslog=true"}, restored.Args)
	assert.Equal(t, CurrentTraefikConfigVersion, setting.Version)
}

func TestTraefikConfigSameArgsAs(t *testing.T) {
	cfg := &TraefikConfig{Args: []string{"traefik", "--log=true", "--log.level=DEBUG"}}

	assert.True(t, cfg.SameArgsAs([]string{"traefik", "--log=true", "--log.level=DEBUG"}))

	// Order is part of the configuration. Traefik takes the last value for a
	// repeated key, so a comparison that ignored order could talk a revert out of
	// restoring a command that behaves differently from the one it is comparing to.
	assert.False(t, cfg.SameArgsAs([]string{"traefik", "--log.level=DEBUG", "--log=true"}))

	assert.False(t, cfg.SameArgsAs([]string{"traefik", "--log=true"}))
	assert.False(t, cfg.SameArgsAs(nil))

	// A service with no arguments and a config with none are the same thing. This
	// is what stops the seeding path from reporting a change on an install where
	// there is nothing to compare.
	assert.True(t, (&TraefikConfig{}).SameArgsAs(nil))
	assert.True(t, (&TraefikConfig{Args: []string{}}).SameArgsAs(nil))
}
