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
	assert.Equal(t, "shop_prod_api_tls_certificate_01234567",
		ObjectName("shop_prod_api", TLSEntry, "certificate", rotation))
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
		"tls-cert": true, "a": true, "a1-b2": true, strings.Repeat("a", 20): true,
		"tls": false, "": false, "-a": false, "a-": false, "A": false, "a_b": false, strings.Repeat("a", 21): false,
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
