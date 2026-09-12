package vlagent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// argsFor returns every value given for one repeated flag, in order.
func argsFor(args []string, name string) []string {
	var out []string
	prefix := name + "="
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			out = append(out, strings.TrimPrefix(a, prefix))
		}
	}
	return out
}

func baseSpec() *loggingmodel.CollectSpec {
	return &loggingmodel.CollectSpec{
		Ingest:  loggingmodel.Endpoint{URL: "http://vlogs:9428/internal/insert"},
		Sources: []loggingmodel.Source{{Kind: loggingmodel.SourceKindApp, Glob: "/var/lib/docker/containers/*/*-json.log"}},
	}
}

func TestConfigureShipsToTheIngestEndpoint(t *testing.T) {
	spec, err := New(&Config{}).Configure(baseSpec())
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"http://vlogs:9428/internal/insert"}, argsFor(spec.Args, "-remoteWrite.url"))
	assert.Equal(t, []string{FormatNative}, argsFor(spec.Args, "-remoteWrite.format"))
	assert.Equal(t, DefaultImage, spec.Image)
}

func TestConfigureMountsTheContainerLogsReadOnly(t *testing.T) {
	spec, err := New(&Config{}).Configure(baseSpec())
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	var found bool
	for _, m := range spec.Mounts {
		if m.Target == ContainersPath {
			found = true
			assert.True(t, m.ReadOnly, "the collector must never be able to write to the log files")
		}
	}
	assert.True(t, found, "container logs are not mounted")
}

// Every -remoteWrite.* array is matched to its url by position, so a
// destination without a credential still needs its slot. Getting this wrong
// sends one target's token to another.
func TestConfigureKeepsPerDestinationArraysAligned(t *testing.T) {
	in := baseSpec()
	in.Forwards = []loggingmodel.ForwardTarget{
		{Name: "no-auth", Format: FormatJSONLine, Endpoint: loggingmodel.Endpoint{URL: "http://a/ingest"}},
		{Name: "with-token", Format: FormatJSONLine, Endpoint: loggingmodel.Endpoint{
			URL: "http://b/ingest", BearerToken: "SECRET",
		}},
	}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	urls := argsFor(spec.Args, "-remoteWrite.url")
	formats := argsFor(spec.Args, "-remoteWrite.format")
	tokens := argsFor(spec.Args, "-remoteWrite.bearerToken")

	assert.Equal(t, []string{"http://vlogs:9428/internal/insert", "http://a/ingest", "http://b/ingest"}, urls)
	assert.Equal(t, []string{FormatNative, FormatJSONLine, FormatJSONLine}, formats)
	// Three slots for three destinations; only the third carries a token.
	assert.Equal(t, []string{"", "", "SECRET"}, tokens)
}

func TestConfigureExcludesTheLoggingStackItself(t *testing.T) {
	in := baseSpec()
	in.Sources[0].Exclude = "/var/lib/docker/containers/deadbeef/*"

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"/var/lib/docker/containers/deadbeef/*"}, argsFor(spec.Args, "-fileCollector.excludeGlob"))
}

// excludeGlob is matched to its glob by position too, so a source with nothing
// to exclude still occupies a slot.
func TestConfigureKeepsGlobAndExcludeAligned(t *testing.T) {
	in := baseSpec()
	in.Sources = []loggingmodel.Source{
		{Kind: loggingmodel.SourceKindApp, Glob: "/a/*.log"},
		{Kind: loggingmodel.SourceKindHivePaaS, Glob: "/b/*.log", Exclude: "/b/skip-*.log"},
	}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{"/a/*.log", "/b/*.log"}, argsFor(spec.Args, "-fileCollector.glob"))
	assert.Equal(t, []string{"", "/b/skip-*.log"}, argsFor(spec.Args, "-fileCollector.excludeGlob"))
}

func TestConfigureRejectsAnEmptySpec(t *testing.T) {
	_, err := New(&Config{}).Configure(&loggingmodel.CollectSpec{
		Ingest: loggingmodel.Endpoint{URL: "http://x"},
	})

	assert.Error(t, err, "collecting nothing is a misconfiguration, not a quiet no-op")
}

func TestConfigureRejectsAMissingIngestURL(t *testing.T) {
	in := baseSpec()
	in.Ingest.URL = ""

	_, err := New(&Config{}).Configure(in)

	assert.Error(t, err)
}
