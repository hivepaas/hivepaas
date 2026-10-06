package base

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestPingArgsAreNotSettable(t *testing.T) {
	for _, key := range []string{"ping", "ping.entrypoint", "entrypoints.ping.address", "PING", "Ping.Entrypoint"} {
		if IsTraefikCmdArgSettable(key) {
			t.Errorf("%q must not be settable: the healthcheck probes it, and a traefik that "+
				"stops answering there is restarted by swarm over and over", key)
		}
	}
	if !IsTraefikCmdArgSettable("log.level") {
		t.Error("log.level should stay settable")
	}
}

// The Access Log option keeps its fields and drops the rest; an argument
// keeping one of its fields is its own, keeping another the operator's.
func TestTheAccessLogOptionOwnsTheFieldsItKeeps(t *testing.T) {
	for _, want := range []string{"--accesslog.format=json", "--accesslog.fields.defaultmode=drop",
		"--accesslog.fields.names.ServiceName=keep", "--accesslog.fields.names.OriginDuration=keep"} {
		if !slices.Contains(TraefikAccessLogArgs, want) {
			t.Errorf("the Access Log option should write %s", want)
		}
	}
	for _, key := range []string{"accesslog", "AccessLog.Format", "accesslog.fields.defaultmode",
		"accesslog.fields.names.servicename", "accesslog.fields.names.ServiceName"} {
		if !IsTraefikAccessLogArg(key) {
			t.Errorf("%q should be the Access Log option's", key)
		}
	}
	for _, key := range []string{"accesslog.fields.names.startutc", "accesslog.fields.headers.defaultmode",
		"accesslog.filepath", "log.level"} {
		if IsTraefikAccessLogArg(key) {
			t.Errorf("%q should be the operator's", key)
		}
	}
}

// A stack starts traefik with the access log the Access Log option would
// write, in that order: an install counts from the same fields a save keeps.
func TestTheStacksWriteTheAccessLogAsTheOptionDoes(t *testing.T) {
	var want strings.Builder
	for _, arg := range TraefikAccessLogArgs {
		want.WriteString(`- "` + arg + `"` + "\n")
	}
	for _, env := range []string{"local", "dev", "release"} {
		b, err := os.ReadFile("../../deployment/" + env + "/hivepaas.yaml")
		if err != nil {
			t.Fatal(err)
		}
		var args strings.Builder
		for line := range strings.Lines(string(b)) {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, `- "--accesslog`) {
				args.WriteString(line + "\n")
			}
		}
		if args.String() != want.String() {
			t.Errorf("deployment/%s/hivepaas.yaml starts traefik with\n%s\nthe Access Log option writes\n%s",
				env, args.String(), want.String())
		}
	}
}
