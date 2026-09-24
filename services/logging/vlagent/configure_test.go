package vlagent

import (
	"encoding/json"
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
		Sources: []*loggingmodel.Source{{Kind: loggingmodel.SourceKindApp, Glob: "/var/lib/docker/containers/*/*-json.log"}},
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
	in.Forwards = []*loggingmodel.ForwardTarget{
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
	tokenFiles := argsFor(spec.Args, "-remoteWrite.bearerTokenFile")

	assert.Equal(t, []string{"http://vlogs:9428/internal/insert", "http://a/ingest", "http://b/ingest"}, urls)
	assert.Equal(t, []string{FormatNative, FormatJSONLine, FormatJSONLine}, formats)
	// Three slots for three destinations; only the third names a token file.
	assert.Equal(t, []string{"", "", SecretsDir + "/remote-write-2-bearer-token"}, tokenFiles)
	assert.Equal(t, []*loggingmodel.Secret{{
		Key: "REMOTE_WRITE_2_BEARER_TOKEN", Path: SecretsDir + "/remote-write-2-bearer-token", Value: "SECRET",
	}}, spec.Secrets)
}

// A credential on the command line is readable by anyone who can read the
// service, so none may appear there - only the file it is written to.
func TestConfigureKeepsCredentialsOffTheCommandLine(t *testing.T) {
	in := baseSpec()
	in.Forwards = []*loggingmodel.ForwardTarget{{
		Name: "company", Format: FormatJSONLine, Endpoint: loggingmodel.Endpoint{
			URL: "http://b/ingest", Username: "shipper", Password: "PASSWORD", BearerToken: "TOKEN",
		},
	}}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	for _, arg := range spec.Args {
		assert.NotContains(t, arg, "PASSWORD")
		assert.NotContains(t, arg, "TOKEN")
	}
	assert.Equal(t, []string{"", "shipper"}, argsFor(spec.Args, "-remoteWrite.basicAuth.username"))
	assert.Equal(t, []string{"", SecretsDir + "/remote-write-1-password"},
		argsFor(spec.Args, "-remoteWrite.basicAuth.passwordFile"))
	assert.Len(t, spec.Secrets, 2)
	assert.True(t, spec.PerNode, "the collector reads the node it runs on")
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
	in.Sources = []*loggingmodel.Source{
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

// vlagent parses this flag as JSON and exits at startup on anything else:
// `cannot parse -fileCollector.extraFields ... cannot parse JSON`, measured
// against v1.52.0. A collector that exits is one swarm restarts forever.
func TestConfigureRendersExtraFieldsAsJSON(t *testing.T) {
	in := baseSpec()
	in.Sources[0].Labels = map[string]string{"hivepaas_source": "hivepaas", "zone": "eu"}

	spec, err := New(&Config{}).Configure(in)
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	fields := argsFor(spec.Args, "-fileCollector.extraFields")
	if len(fields) != 1 {
		t.Fatalf("want one slot, got %d", len(fields))
	}
	assert.Equal(t, `{"hivepaas_source":"hivepaas","zone":"eu"}`, fields[0],
		"JSON, with keys in a stable order")

	var parsed map[string]string
	assert.NoError(t, json.Unmarshal([]byte(fields[0]), &parsed), "vlagent must be able to parse it")
	assert.Equal(t, in.Sources[0].Labels, parsed)
}

func TestConfigureLeavesExtraFieldsEmptyWithoutLabels(t *testing.T) {
	spec, err := New(&Config{}).Configure(baseSpec())
	if err != nil {
		t.Fatalf("Configure: %v", err)
	}

	assert.Equal(t, []string{""}, argsFor(spec.Args, "-fileCollector.extraFields"))
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

func TestAttrFieldIsWhereADaemonLabelLands(t *testing.T) {
	// Measured: json-file writes the labels option's values into the line's
	// attrs object, and vlagent flattens it with this prefix.
	assert.Equal(t, "attrs.hivepaas.app.id", AttrField("hivepaas.app.id"))
}
