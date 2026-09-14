package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// storedLogging returns a setting shaped like a database row: Data filled in,
// no parsed cache. Reading back through the same *Setting that SetData was
// called on would return the cached struct without ever touching Data, so a
// round-trip through it proves nothing about what persists.
func storedLogging(t *testing.T, data *entity.LoggingSettings) *entity.Setting {
	t.Helper()

	s := &entity.Setting{ID: "s1", Type: base.SettingTypeLogging}
	if err := s.SetData(data); err != nil {
		t.Fatalf("SetData: %v", err)
	}
	return &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
}

func TestLoggingSurvivesPersistence(t *testing.T) {
	stored := storedLogging(t, &entity.LoggingSettings{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true, HivePaaS: true},
		Collector: entity.LoggingCollector{
			Type:    base.LoggingCollectorTypeVlagent,
			Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type:    base.LoggingBackendTypeVictoriaLogs,
			Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{
				Volume: entity.ObjectID{ID: "vol-1"},
			},
		},
		Forwards: []entity.LoggingForward{{
			Name:     "siem",
			Format:   "jsonline",
			Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
		}},
	})

	got, err := stored.AsLoggingSettings()
	if err != nil {
		t.Fatalf("AsLogging: %v", err)
	}

	assert.True(t, got.Enabled)
	assert.True(t, got.Sources.Apps)
	assert.True(t, got.Sources.HivePaaS)
	assert.False(t, got.Sources.Nodes)
	assert.Equal(t, base.LoggingCollectorTypeVlagent, got.Collector.Type)
	assert.True(t, got.Collector.Managed)
	assert.Equal(t, base.LoggingBackendTypeVictoriaLogs, got.Backend.Type)
	if got.Backend.VictoriaLogs == nil {
		t.Fatal("VictoriaLogs block did not survive")
	}
	assert.Equal(t, "vol-1", got.Backend.VictoriaLogs.Volume.ID)
	if len(got.Forwards) != 1 {
		t.Fatalf("want one forward, got %d", len(got.Forwards))
	}
	assert.Equal(t, "siem", got.Forwards[0].Name)
	assert.Equal(t, "https://siem.example/ingest", got.Forwards[0].Endpoint.URL)
}

// The volume the backend writes to is a real reference: it must be refused if it
// does not exist, and it must not be deletable while logging uses it. This is
// what VerifyingRefIDs acts on, unlike ClusterVolume where it is always empty.
func TestLoggingReferencesItsDataVolume(t *testing.T) {
	l := &entity.LoggingSettings{Backend: entity.LoggingBackend{
		VictoriaLogs: &entity.LoggingVictoriaLogs{Volume: entity.ObjectID{ID: "vol-9"}},
	}}

	assert.Equal(t, []string{"vol-9"}, l.GetRefObjectIDs().RefSettingIDs)
}

func TestLoggingReferencesNothingWithoutAManagedBackend(t *testing.T) {
	l := &entity.LoggingSettings{Backend: entity.LoggingBackend{Type: base.LoggingBackendTypeVictoriaLogs}}

	assert.Empty(t, l.GetRefObjectIDs().RefSettingIDs)
}

func TestLoggingIsDisabledByDefault(t *testing.T) {
	stored := storedLogging(t, &entity.LoggingSettings{})

	got, err := stored.AsLoggingSettings()
	if err != nil {
		t.Fatalf("AsLogging: %v", err)
	}

	assert.False(t, got.Enabled, "logging must not deploy anything until it is switched on")
}

// Credentials are stored encrypted and read back as plaintext, which is what
// lets the service layer hand services/logging plain strings.
func TestLoggingDecryptsEveryEndpoint(t *testing.T) {
	l := &entity.LoggingSettings{
		Backend:  entity.LoggingBackend{Ingest: &entity.LoggingEndpoint{URL: "https://a"}},
		Forwards: []entity.LoggingForward{{Name: "f", Endpoint: entity.LoggingEndpoint{URL: "https://b"}}},
	}
	l.Backend.Ingest.BearerToken.Set("INGEST-TOKEN")
	l.Forwards[0].Endpoint.BearerToken.Set("FORWARD-TOKEN")

	if err := l.Decrypt(); err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	ingest, err := l.Backend.Ingest.BearerToken.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "INGEST-TOKEN", ingest)

	forward, err := l.Forwards[0].Endpoint.BearerToken.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "FORWARD-TOKEN", forward)
}
