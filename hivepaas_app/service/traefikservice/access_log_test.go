package traefikservice

import (
	"regexp"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
)

func traefikSpec(args ...string) *swarm.ServiceSpec {
	return &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Args: args, Labels: map[string]string{"keep": "me"}},
		LogDriver: &swarm.Driver{Name: "json-file",
			Options: map[string]string{"max-size": "50m", "labels": "other.label"}},
	}}
}

// The proxy's lines get its identity, its other labels and options kept; a
// second time changes nothing, so the spec stays the same.
func TestWithAccessLogIdentity(t *testing.T) {
	spec := traefikSpec("--accesslog=true", "--accesslog.format=json")
	assert.Equal(t, AccessLogUnlabelled, AccessLogReadiness(spec))

	assert.True(t, WithAccessLogIdentity(spec))
	cs := spec.TaskTemplate.ContainerSpec
	assert.Equal(t, "traefik", cs.Labels["hivepaas.component"])
	assert.Equal(t, "me", cs.Labels["keep"])
	assert.Equal(t, "hivepaas.component,other.label", spec.TaskTemplate.LogDriver.Options["labels"])
	assert.Equal(t, "50m", spec.TaskTemplate.LogDriver.Options["max-size"])
	assert.Equal(t, AccessLogNotReadyReason(""), AccessLogReadiness(spec))

	assert.False(t, WithAccessLogIdentity(spec), "already there")

	noDriver := traefikSpec()
	noDriver.TaskTemplate.LogDriver = nil
	assert.True(t, WithAccessLogIdentity(noDriver))
	assert.Equal(t, "json-file", noDriver.TaskTemplate.LogDriver.Name)
	assert.Equal(t, "hivepaas.component", noDriver.TaskTemplate.LogDriver.Options["labels"])

	other := traefikSpec()
	other.TaskTemplate.LogDriver = &swarm.Driver{Name: "syslog"}
	WithAccessLogIdentity(other)
	assert.Equal(t, "syslog", other.TaskTemplate.LogDriver.Name, "another driver is the operator's")
}

// Readiness says what is missing, first things first.
func TestAccessLogReadiness(t *testing.T) {
	ready := traefikSpec("traefik", "--accesslog=true", "--accesslog.format=json")
	WithAccessLogIdentity(ready)
	assert.Equal(t, AccessLogNotReadyReason(""), AccessLogReadiness(ready))

	off := traefikSpec("--log.level=INFO")
	WithAccessLogIdentity(off)
	assert.Equal(t, AccessLogOff, AccessLogReadiness(off))

	clf := traefikSpec("--accesslog=true")
	WithAccessLogIdentity(clf)
	assert.Equal(t, AccessLogNotJSON, AccessLogReadiness(clf))

	assert.Equal(t, AccessLogOff, AccessLogReadiness(nil))
}

// The access log's arguments become the release's - the fields it keeps among
// them - where they were; the operator's stay, and so does an access log off.
func TestWithAccessLogArgs(t *testing.T) {
	release := func(before []string, after ...string) []string {
		return append(append(before, base.TraefikAccessLogArgs...), after...)
	}

	spec := traefikSpec("traefik", "--log=true", "--accesslog=true", "--accesslog.format=json",
		"--accesslog.fields.queryparameters.defaultmode=drop", "--providers.docker=true",
		"--accesslog.fields.names.StartUTC=keep", "--accesslog.fields.names.ServiceName=drop")
	assert.True(t, WithAccessLogArgs(spec))
	assert.Equal(t, release([]string{"traefik", "--log=true"}, "--providers.docker=true",
		"--accesslog.fields.names.StartUTC=keep"), spec.TaskTemplate.ContainerSpec.Args,
		"a field of the operator's kept, none of the option's dropped")
	assert.False(t, WithAccessLogArgs(spec), "already there")

	clf := traefikSpec("traefik", "--accesslog")
	WithAccessLogIdentity(clf)
	assert.Equal(t, AccessLogNotJSON, AccessLogReadiness(clf))
	assert.True(t, WithAccessLogArgs(clf))
	assert.Equal(t, release([]string{"traefik"}), clf.TaskTemplate.ContainerSpec.Args)
	assert.Equal(t, AccessLogNotReadyReason(""), AccessLogReadiness(clf))

	for _, args := range [][]string{
		{"traefik", "--accesslog=false", "--accesslog.format=json"},
		{"traefik", "--accesslog=true", "--accesslog=false"},
		{"traefik", "--log=true"},
	} {
		off := traefikSpec(args...)
		assert.False(t, WithAccessLogArgs(off), "%v", args)
		assert.Equal(t, args, off.TaskTemplate.ContainerSpec.Args)
	}

	assert.False(t, WithAccessLogArgs(&swarm.ServiceSpec{}))
}

// Every field an app's HTTP numbers and request load are counted from is one
// the proxy writes: the Access Log option has Traefik drop the others.
func TestTheAccessLogKeepsTheFieldsCountedFrom(t *testing.T) {
	match := []loggingmodel.FieldMatch{{Field: "attrs.hivepaas.component", Value: "traefik"}}
	start := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	stats, err := victorialogs.BuildHTTPStatsQueries(&loggingmodel.HTTPStatsReq{
		Match: match, ServicePattern: "^svc-01k6a-[0-9]+@swarm$", Start: start, End: start.Add(time.Hour),
		Step: time.Minute, TopPaths: 20, TopReplicas: 10,
	})
	assert.NoError(t, err)
	load, err := victorialogs.BuildRequestLoadQuery(&loggingmodel.RequestLoadReq{
		Match: match, AppIDs: []string{"01K6A"}, Start: start, End: start.Add(time.Minute),
	})
	assert.NoError(t, err)

	taken := regexp.MustCompile(`extract "\\"(\w+)\\":<`)
	var fields []string
	for _, query := range []string{stats.Series, stats.Totals, stats.Paths, stats.Replicas, load} {
		for _, m := range taken.FindAllStringSubmatch(query, -1) {
			fields = append(fields, m[1])
		}
	}
	assert.Subset(t, fields, []string{"ServiceName", "ServiceURL", "RequestMethod", "RequestPath",
		"DownstreamStatus", "OriginStatus", "Duration", "OriginDuration"}, "each query read")
	assert.Subset(t, base.TraefikAccessLogFields, fields)
}
