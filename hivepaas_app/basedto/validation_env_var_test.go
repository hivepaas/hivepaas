package basedto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"
)

func TestValidateEnvVarsReq(t *testing.T) {
	tests := []struct {
		name      string
		envVars   []*EnvVarReq
		expectErr bool
		errField  string
		errKey    string
	}{
		{
			name: "Valid Env Names",
			envVars: []*EnvVarReq{
				{Key: "DATABASE_URL", Value: "postgres://..."},
				{Key: "_PORT_NUMBER", Value: "8080"},
				{Key: "app_env", Value: "production"},
			},
			expectErr: false,
		},
		{
			name: "Invalid name starting with number",
			envVars: []*EnvVarReq{
				{Key: "123PORT", Value: "8080"},
			},
			expectErr: true,
			errField:  "buildtimeEnvVars[0].key",
			errKey:    "ERR_VLD_NAME_INVALID",
		},
		{
			name: "Invalid name containing hyphen",
			envVars: []*EnvVarReq{
				{Key: "DB-PASSWORD", Value: "secret"},
			},
			expectErr: true,
			errField:  "buildtimeEnvVars[0].key",
			errKey:    "ERR_VLD_NAME_INVALID",
		},
		{
			name: "Duplicate keys",
			envVars: []*EnvVarReq{
				{Key: "PORT", Value: "80"},
				{Key: "PORT", Value: "443"},
			},
			expectErr: true,
			errField:  "buildtimeEnvVars",
			errKey:    "ERR_VLD_VALUES_NON_UNIQUE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validators := ValidateEnvVarsReq(tt.envVars, "buildtimeEnvVars")
			errs := vld.Validate(validators...)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				// check if the specific error has the correct field path and error key
				found := false
				for _, err := range errs {
					fieldPath := ""
					f := err.Field()
					if f != nil {
						fieldPath = f.PathString(true, "/")
					}
					customKey := ""
					if ck := err.CustomKey(); ck != nil {
						customKey = ck.(string)
					}
					if fieldPath == tt.errField && customKey == tt.errKey {
						found = true
						break
					}
				}
				assert.True(t, found,
					"Expected error on field %s with key %s, but not found. Errors: %v",
					tt.errField, tt.errKey, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
