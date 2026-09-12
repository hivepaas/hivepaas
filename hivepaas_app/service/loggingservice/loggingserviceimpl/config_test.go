package loggingserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/logging"
)

// The service layer must hand services/logging plain strings. An EncryptedField
// arriving undecrypted would reach vlagent's command line as ciphertext and the
// forward would fail authentication with nothing saying why.
func TestToEndpointDecrypts(t *testing.T) {
	ep := &entity.LoggingEndpoint{URL: "https://siem.example/ingest", Username: "u"}
	// Set takes plaintext and does not return an error; it detects whether the
	// value is already ciphertext. GetPlain then short-circuits on the
	// plaintext it holds, so this needs no configured data key.
	ep.BearerToken.Set("SECRET")

	got, err := toEndpoint(ep)
	if err != nil {
		t.Fatalf("toEndpoint: %v", err)
	}

	assert.Equal(t, "https://siem.example/ingest", got.URL)
	assert.Equal(t, "u", got.Username)
	assert.Equal(t, "SECRET", got.BearerToken)
}

func TestToEndpointHandlesNil(t *testing.T) {
	got, err := toEndpoint(nil)

	assert.NoError(t, err)
	assert.Empty(t, got.URL)
}

func TestBuildCollectSpecIncludesOnlySelectedSources(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{Sources: entity.LoggingSources{Apps: true, TraefikAccess: true}}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	kinds := map[logging.SourceKind]bool{}
	for _, src := range spec.Sources {
		kinds[src.Kind] = true
	}
	assert.True(t, kinds[logging.SourceKindApp])
	assert.True(t, kinds[logging.SourceKindTraefikAccess])
	assert.False(t, kinds[logging.SourceKindHivePaaS], "HivePaaS logs are opt-in")
	assert.False(t, kinds[logging.SourceKindNode])
}

// The stack writes logs of its own, and collecting them feeds the collector its
// own output. Excluding both is required, not an optimisation.
func TestBuildCollectSpecExcludesTheLoggingStack(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{Sources: entity.LoggingSources{Apps: true}}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	var appSource *logging.Source
	for i := range spec.Sources {
		if spec.Sources[i].Kind == logging.SourceKindApp {
			appSource = &spec.Sources[i]
		}
	}
	if appSource == nil {
		t.Fatal("no app source")
	}
	assert.NotEmpty(t, appSource.Exclude)
	assert.Contains(t, appSource.Exclude, ServiceNameCollector)
}

func TestBuildCollectSpecCarriesForwards(t *testing.T) {
	s := &service{}
	cfg := &entity.Logging{
		Sources: entity.LoggingSources{Apps: true},
		Forwards: []entity.LoggingForward{{
			Name:     "siem",
			Format:   "jsonline",
			Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
		}},
	}

	spec, err := s.buildCollectSpec(cfg, "http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	if len(spec.Forwards) != 1 {
		t.Fatalf("want one forward, got %d", len(spec.Forwards))
	}
	assert.Equal(t, "siem", spec.Forwards[0].Name)
	assert.Equal(t, "https://siem.example/ingest", spec.Forwards[0].Endpoint.URL)
	assert.Equal(t, "http://vlogs:9428/internal/insert", spec.Ingest.URL)
}

func TestBuildCollectSpecRefusesWithNoSources(t *testing.T) {
	s := &service{}

	_, err := s.buildCollectSpec(&entity.Logging{}, "http://vlogs:9428/internal/insert")

	assert.Error(t, err)
}

// A source glob must name the docker log directory, or the collector reads
// nothing and reports no error.
func TestAppSourceGlobPointsAtDockerLogs(t *testing.T) {
	s := &service{}

	spec, err := s.buildCollectSpec(&entity.Logging{Sources: entity.LoggingSources{Apps: true}},
		"http://vlogs:9428/internal/insert")
	if err != nil {
		t.Fatalf("buildCollectSpec: %v", err)
	}

	assert.True(t, strings.HasPrefix(spec.Sources[0].Glob, "/var/lib/docker/containers/"))
}
