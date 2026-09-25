package settingmountservice

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rotation = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestObjectNamesFollowTheSecretsConvention(t *testing.T) {
	assert.Equal(t, "shop_prod_api_mount_tls-cert_privatekey_01234567",
		ObjectName("Shop_Prod_API", "tls-cert", "privateKey", rotation))
	// "tls" is an entry like any other: nothing mounts under a name of its own.
	assert.Equal(t, "shop_prod_api_mount_tls_certificate_01234567",
		ObjectName("shop_prod_api", "tls", "certificate", rotation))
}

// Docker caps names at 64; the GlobalKey is cut and a hash of it keeps two long
// keys that share a beginning apart.
func TestALongNameIsShortenedAndStaysUnique(t *testing.T) {
	long := strings.Repeat("project-with-a-long-name_", 3) + "env_app"
	a := ObjectName(long+"-a", "abcdefghij0123456789", "caCertificate", rotation)
	b := ObjectName(long+"-b", "abcdefghij0123456789", "caCertificate", rotation)
	assert.LessOrEqual(t, len(a), 64)
	assert.LessOrEqual(t, len(b), 64)
	assert.NotEqual(t, a, b)
	assert.True(t, strings.HasSuffix(a, "_mount_abcdefghij0123456789_cacertificate_01234567"), a)
}

func TestEntryKeys(t *testing.T) {
	for key, want := range map[string]bool{
		"tls-cert": true, "tls": true, "a": true, "a1-b2": true, strings.Repeat("a", 20): true,
		"": false, "-a": false, "a-": false, "A": false, "a_b": false, strings.Repeat("a", 21): false,
	} {
		assert.Equal(t, want, ValidEntryKey(key), key)
	}
}

func TestLabels(t *testing.T) {
	assert.Equal(t, map[string]string{
		"hivepaas.app.id":             "app_1",
		"hivepaas.settingMount.entry": "tls-cert",
		"hivepaas.settingMount.part":  "privateKey",
	}, Labels("app_1", "tls-cert", "privateKey"))
}

func TestEntryKeyOfATemplateSetting(t *testing.T) {
	for name, want := range map[string]string{
		"DB_PASSWORD": "db-password", "config.json": "config-json", "nginx-conf": "nginx-conf",
		"__A__": "a", strings.Repeat("x", 30): strings.Repeat("x", 20),
	} {
		assert.Equal(t, want, EntryKeyFor(name), name)
		assert.True(t, ValidEntryKey(EntryKeyFor(name)), name)
	}
}
