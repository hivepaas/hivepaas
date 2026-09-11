package traefiksettingsuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
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

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{
		User: &entity.User{ID: "admin-user", Username: "admin"},
	}}
}

func TestRecordTraefikSettingsUpdateFilesUnderTheInstall(t *testing.T) {
	audit := &spyAuditService{}
	uc := &UC{auditService: audit}

	err := uc.recordTraefikSettingsUpdate(context.Background(), nil, adminAuth(),
		auditSectionConfigOptions, auditdetail.New().Set("commandChanged", true))

	assert.NoError(t, err)
	assert.Len(t, audit.entries, 1)
	entry := audit.entries[0]
	// The hivepaas scope, not the traefik app's: this is the install's own log.
	assert.Equal(t, base.ObjectScopeHivepaas, entry.Scope)
	assert.Empty(t, entry.ObjectID)
	assert.Equal(t, base.AuditLogTypeHivePaaSSettingsUpdate, entry.Type)
	assert.Equal(t, auditSectionConfigOptions, entry.Section)
	assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
	assert.Equal(t, "traefik config options", entry.ResName)
	assert.Contains(t, entry.Detail, `"commandChanged":true`)
}

// The section is what tells these entries from the HivePaaS pages they share a
// type with, so it has to stay distinct from theirs.
func TestTraefikAuditSectionsAreNamedApartFromTheHivePaaSPages(t *testing.T) {
	assert.NotEqual(t, auditSectionConfigOptions, auditSectionServiceSettings)
	for _, section := range []string{auditSectionConfigOptions, auditSectionServiceSettings} {
		assert.NotEqual(t, "routing", section)
		assert.NotEqual(t, "service", section)
		assert.NotEmpty(t, auditResNames[section], "a section has to have a name to read under")
	}
}

// An entry naming nobody answers none of the questions the entry exists for.
func TestRecordTraefikSettingsUpdateRefusesAnUnattributedChange(t *testing.T) {
	audit := &spyAuditService{}
	uc := &UC{auditService: audit}

	err := uc.recordTraefikSettingsUpdate(context.Background(), nil, nil,
		auditSectionServiceSettings, nil)

	assert.Error(t, err)
	assert.Empty(t, audit.entries)
}

// A change that cannot be recorded has to fail, because the caller applies it
// only if this returns nil.
func TestRecordTraefikSettingsUpdateFailsWhenTheStoreIsDown(t *testing.T) {
	audit := &spyAuditService{err: errors.New("audit store is down")}
	uc := &UC{auditService: audit}

	err := uc.recordTraefikSettingsUpdate(context.Background(), nil, adminAuth(),
		auditSectionServiceSettings, nil)

	assert.Error(t, err)
}

func TestChangedCommandArgsNamesTheFlagsAndNotTheirValues(t *testing.T) {
	tests := []struct {
		name   string
		before []string
		after  []string
		want   []string
	}{
		{
			name:   "a flag added",
			before: []string{"traefik", "--log=true"},
			after:  []string{"traefik", "--log=true", "--api.insecure=true"},
			want:   []string{"api.insecure"},
		},
		{
			name:   "a flag removed",
			before: []string{"traefik", "--log=true", "--accesslog=true"},
			after:  []string{"traefik", "--log=true"},
			want:   []string{"accesslog"},
		},
		{
			name:   "a value changed",
			before: []string{"traefik", "--log.level=INFO"},
			after:  []string{"traefik", "--log.level=DEBUG"},
			want:   []string{"log.level"},
		},
		{
			name:   "nothing moved",
			before: []string{"traefik", "--log=true"},
			after:  []string{"traefik", "--log=true"},
			want:   []string{},
		},
		{
			// Order is not reported: the entry says the command changed anyway,
			// and a reorder has no flag to name.
			name:   "only the order moved",
			before: []string{"traefik", "--log=true", "--accesslog=true"},
			after:  []string{"traefik", "--accesslog=true", "--log=true"},
			want:   []string{},
		},
		{
			// The one traefik runs under is the last, so that is the one compared.
			name:   "a repeated flag is judged by its last value",
			before: []string{"traefik", "--log.level=DEBUG", "--log.level=INFO"},
			after:  []string{"traefik", "--log.level=INFO"},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, changedCommandArgs(tt.before, tt.after))
		})
	}
}

// The values are what must never reach the row: a command line is free text, and
// a credential written into one would otherwise outlive the service it came from.
func TestChangedCommandArgsCarriesNoValues(t *testing.T) {
	changed := changedCommandArgs(
		[]string{"traefik"},
		[]string{"traefik", "--certificatesresolvers.le.acme.email=ops@example.com"},
	)

	assert.Equal(t, []string{"certificatesresolvers.le.acme.email"}, changed)
	for _, key := range changed {
		assert.NotContains(t, key, "ops@example.com")
	}
}
