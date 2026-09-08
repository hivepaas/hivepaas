package entity

import (
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentTraefikConfigVersion = 1
)

var _ = registerSettingParser(base.SettingTypeTraefikConfig, &traefikConfigParser{})

type traefikConfigParser struct {
}

func (s *traefikConfigParser) New() SettingData {
	return &TraefikConfig{}
}

// TraefikConfig is the record of traefik's startup command as HivePaaS last
// applied it.
//
// Traefik's own state lives in the swarm service spec, and that stays the source
// of truth for what is running: GetConfigOptions reads it, and an update rebuilds
// the argument list from it so an operator who edited the service by hand is not
// silently overwritten. This row exists for what the spec cannot provide - a
// version to lock updates against, an audit trail, and something for confirm-or-
// revert to hang a snapshot on. Drift is self-healing for that reason: the next
// update reads the live spec and writes the result back here.
//
// Args is the whole list as applied, unsettable system arguments included, so a
// revert restores exactly what was running rather than a reconstruction of it.
type TraefikConfig struct {
	Args []string `json:"args"`
}

func (s *TraefikConfig) GetType() base.SettingType {
	return base.SettingTypeTraefikConfig
}

func (s *TraefikConfig) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *TraefikConfig) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// SameArgsAs reports whether the two command lines are the one traefik is
// already running.
//
// Order counts. Traefik takes the last value for a repeated key, so two lists
// with the same elements in a different order are not necessarily the same
// configuration, and a comparison that ignored order could talk a revert out of
// work it needed to do.
func (s *TraefikConfig) SameArgsAs(args []string) bool {
	return slices.Equal(s.Args, args)
}

func (s *Setting) AsTraefikConfig() (*TraefikConfig, error) {
	return parseSettingAs[*TraefikConfig](s)
}

func (s *Setting) MustAsTraefikConfig() *TraefikConfig {
	return gofn.Must(s.AsTraefikConfig())
}
