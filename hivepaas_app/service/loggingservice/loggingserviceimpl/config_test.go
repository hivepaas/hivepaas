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

// Apps and HivePaaS share one directory of container ids, so they are one
// source. Two identical globs collect the file once and apply only the first
// slot's extra fields - measured against vlagent v1.52.0 - so asking for two
// would quietly drop the second source's label.
func TestBuildCollectSpecCollectsContainerLogsAsOneSource(t *testing.T) {
	s := &service{}

	for _, tc := range []struct {
		name    string
		sources entity.LoggingSources
		want    logging.SourceKind
	}{
		{"apps only", entity.LoggingSources{Apps: true}, logging.SourceKindApp},
		{"hivepaas only", entity.LoggingSources{HivePaaS: true}, logging.SourceKindHivePaaS},
		{"both", entity.LoggingSources{Apps: true, HivePaaS: true}, logging.SourceKindApp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := s.buildCollectSpec(&entity.Logging{Sources: tc.sources},
				"http://vlogs:9428/internal/insert")
			if err != nil {
				t.Fatalf("buildCollectSpec: %v", err)
			}

			if len(spec.Sources) != 1 {
				t.Fatalf("want one source, got %d", len(spec.Sources))
			}
			assert.Equal(t, tc.want, spec.Sources[0].Kind)
			assert.Equal(t, dockerContainersGlob, spec.Sources[0].Glob)
			assert.Empty(t, spec.Sources[0].Labels,
				"the lines are a mix of apps and HivePaaS; labeling them all one way would be a lie")
		})
	}
}

// A glob pointing outside the collector's only mount matches nothing and
// reports no error, which is worse than not offering the source.
func TestBuildCollectSpecCollectsNothingForSourcesItCannotRead(t *testing.T) {
	s := &service{}

	_, err := s.buildCollectSpec(
		&entity.Logging{Sources: entity.LoggingSources{TraefikAccess: true, Nodes: true}},
		"http://vlogs:9428/internal/insert")

	assert.ErrorIs(t, err, logging.ErrNoSources)
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
