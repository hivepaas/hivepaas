package registryauthdto

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"
)

func passwordErrors(password string) []string {
	req := &RegistryAuthBaseReq{Name: "gar", Address: "asia-southeast1-docker.pkg.dev",
		Username: "_json_key_base64", Password: password}
	var out []string
	for _, err := range vld.Validate(req.validate("auth")...) {
		if f := err.Field(); f != nil {
			out = append(out, f.PathString(true, "."))
		}
	}
	return out
}

// A Google Artifact Registry credential that does not expire is a service
// account's JSON key, in base64: some 3 KB, which the password takes.
func TestARegistryPasswordTakesAJSONKey(t *testing.T) {
	key := `{"type":"service_account","private_key":"` + strings.Repeat("A", 1700) + `"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(key))
	assert.Greater(t, len(encoded), 2000)
	assert.Empty(t, passwordErrors(encoded))

	assert.Contains(t, passwordErrors(strings.Repeat("x", 8*1024+1)), "auth.password", "past 8 KB")
}
