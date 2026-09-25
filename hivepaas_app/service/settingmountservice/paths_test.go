package settingmountservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	for path, want := range map[string]bool{
		"/etc/app/tls/cert.pem": true, "/run/secrets/app.pem": true, "/run/secrets/tlsx": true,
		"/run/secrets/tls": true, "/run/secrets/tls/cert.pem": true,
		"": false, "/": false, "etc/app": false, "/etc/../app": false, "/etc/app/": false,
	} {
		assert.Equal(t, want, ValidPath(path), path)
	}
}

func TestTargetsOfOrdinaryReferences(t *testing.T) {
	assert.Equal(t, "/run/secrets/db_password", SecretTarget("db_password"))
	assert.Equal(t, "/etc/app/key", SecretTarget("/etc/app/key"))
	assert.Equal(t, "/app.conf", ConfigTarget("app.conf"))
	assert.Equal(t, "/etc/app.conf", ConfigTarget("/etc/app.conf"))
}
