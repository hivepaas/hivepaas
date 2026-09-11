package traefikuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

// spyAuditService keeps what was recorded, so a test can assert on the record
// itself rather than only on the answer the caller got.
type spyAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (s *spyAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	s.entries = append(s.entries, entry)
	return s.err
}

// spyTraefikService embeds the interface so only the methods these endpoints
// reach need a body; anything else would panic loudly rather than pass.
type spyTraefikService struct {
	traefikservice.Service
	restarts int
	reloads  int
	resets   int
}

func (s *spyTraefikService) RestartTraefikSwarmService(_ context.Context) error {
	s.restarts++
	return nil
}

func (s *spyTraefikService) ReloadTraefikConfig(_ context.Context, _ bool) error {
	s.reloads++
	return nil
}

func (s *spyTraefikService) ResetTraefikConfig(_ context.Context) error {
	s.resets++
	return nil
}

func newUCTest() (*UC, *spyAuditService, *spyTraefikService) {
	audit := &spyAuditService{}
	traefik := &spyTraefikService{}
	return &UC{auditService: audit, traefikService: traefik}, audit, traefik
}

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: "admin-user", Username: "admin"},
	}}
}

func TestTraefikActionsAreRecorded(t *testing.T) {
	tests := []struct {
		name    string
		call    func(uc *UC) error
		section string
		done    func(traefik *spyTraefikService) int
	}{
		{
			name: "restart",
			call: func(uc *UC) error {
				_, err := uc.RestartTraefik(context.Background(), adminAuth(),
					&traefikdto.RestartTraefikReq{})
				return err
			},
			section: auditSectionRestart,
			done:    func(traefik *spyTraefikService) int { return traefik.restarts },
		},
		{
			name: "config reload",
			call: func(uc *UC) error {
				_, err := uc.ReloadTraefikConfig(context.Background(), adminAuth(),
					&traefikdto.ReloadTraefikConfigReq{})
				return err
			},
			section: auditSectionConfigReload,
			done:    func(traefik *spyTraefikService) int { return traefik.reloads },
		},
		{
			name: "config reset",
			call: func(uc *UC) error {
				_, err := uc.ResetTraefikConfig(context.Background(), adminAuth(),
					&traefikdto.ResetTraefikConfigReq{})
				return err
			},
			section: auditSectionConfigReset,
			done:    func(traefik *spyTraefikService) int { return traefik.resets },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, audit, traefik := newUCTest()

			assert.NoError(t, tt.call(uc))
			assert.Equal(t, 1, tt.done(traefik))

			assert.Len(t, audit.entries, 1)
			entry := audit.entries[0]
			assert.Equal(t, base.AuditLogTypeHivePaaSAction, entry.Type)
			// The install's own log, not the traefik app's.
			assert.Equal(t, base.ObjectScopeHivepaas, entry.Scope)
			assert.Empty(t, entry.ObjectID)
			assert.Equal(t, base.AuditLogSourceAPIAction, entry.Source)
			assert.Equal(t, tt.section, entry.Section)
			assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
			assert.Equal(t, "admin-user", entry.Auth.User.ID)
		})

		t.Run(tt.name+" does not happen when it cannot be recorded", func(t *testing.T) {
			uc, audit, traefik := newUCTest()
			audit.err = errors.New("audit store is down")

			assert.Error(t, tt.call(uc))
			assert.Equal(t, 0, tt.done(traefik),
				"an action that could not be recorded must not be carried out")
		})
	}
}

// The sections share a type with what the HivePaaS app does, so they have to stay
// distinct from its "restart" and "config-reload".
func TestTraefikAuditSectionsAreNamedApartFromTheHivePaaSApp(t *testing.T) {
	sections := []string{auditSectionRestart, auditSectionConfigReload, auditSectionConfigReset}
	for _, section := range sections {
		assert.NotEqual(t, "restart", section)
		assert.NotEqual(t, "config-reload", section)
		assert.NotEqual(t, "version-update", section)
	}
	assert.Len(t, map[string]bool{
		sections[0]: true, sections[1]: true, sections[2]: true,
	}, len(sections), "each action needs a section of its own")
}

// An entry naming nobody answers none of the questions the entry exists for.
func TestRecordTraefikActionRefusesAnUnattributedAction(t *testing.T) {
	uc, audit, _ := newUCTest()

	err := uc.recordTraefikAction(context.Background(), nil, nil, auditSectionRestart)

	assert.Error(t, err)
	assert.Empty(t, audit.entries)
}
