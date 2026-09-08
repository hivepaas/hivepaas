package base

import "testing"

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
