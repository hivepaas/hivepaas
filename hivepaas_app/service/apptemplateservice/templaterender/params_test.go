package templaterender

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

func intPtr(v int) *int { return &v }

func testParams() []*templatemodel.Parameter {
	return []*templatemodel.Parameter{
		{Name: "dbName", Type: templatemodel.ParamTypeString, Default: "app", Pattern: "^[a-z_]+$",
			MaxLength: intPtr(8)},
		{Name: "password", Type: templatemodel.ParamTypeSecret,
			Generate: &templatemodel.Generate{Length: 24}},
		{Name: "token", Type: templatemodel.ParamTypeSecret, Optional: true},
		{Name: "workers", Type: templatemodel.ParamTypeInt, Default: 4, Min: 1, Max: 64},
		{Name: "memoryLimit", Type: templatemodel.ParamTypeSize, Default: "512MB", Min: "128MB"},
		{Name: "debug", Type: templatemodel.ParamTypeBool, Default: false},
		{Name: "mode", Type: templatemodel.ParamTypeSelect, Default: "fast",
			Options: []*templatemodel.SelectOption{{Value: "fast", Title: "Fast"}, {Value: "safe", Title: "Safe"}}},
		{Name: "dataVolume", Type: templatemodel.ParamTypeVolume},
	}
}

func paramDetail(t *testing.T, err error) string {
	t.Helper()
	var hpErr hperrors.HPError
	if !errors.As(err, &hpErr) {
		t.Fatalf("expected an hperrors.HPError, got %T: %v", err, err)
	}
	return hpErr.Build("en").Detail
}

func TestResolveParamsFillsDefaultsAndGeneratesSecrets(t *testing.T) {
	values, err := ResolveParams(testParams(), map[string]any{"dataVolume": "vol-1"})

	assert.NoError(t, err)
	assert.Equal(t, "app", values["dbName"].Value)
	assert.Equal(t, int64(4), values["workers"].Value)
	assert.Equal(t, unit.DataSize(512<<20), values["memoryLimit"].Value)
	assert.Equal(t, false, values["debug"].Value)
	assert.Equal(t, "fast", values["mode"].Value)
	assert.Equal(t, "vol-1", values["dataVolume"].Value)
	assert.Nil(t, values["token"].Value, "an optional parameter nobody set stays unset")
	assert.Equal(t, "", values["token"].Text())

	password := values["password"]
	assert.True(t, password.Generated)
	assert.Regexp(t, regexp.MustCompile(`^[A-Za-z0-9]{24}$`), password.Text())

	again, err := ResolveParams(testParams(), map[string]any{"dataVolume": "vol-1"})
	assert.NoError(t, err)
	assert.NotEqual(t, password.Text(), again["password"].Text(), "every generated secret is new")
}

func TestResolveParamsConvertsInput(t *testing.T) {
	values, err := ResolveParams(testParams(), map[string]any{
		"dbName":      "shop",
		"password":    "given-password",
		"workers":     float64(16), // what a JSON body decodes to
		"memoryLimit": "1GB",
		"debug":       "true", // what the CLI passes
		"mode":        "safe",
		"dataVolume":  "vol-2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "shop", values["dbName"].Value)
	assert.Equal(t, "given-password", values["password"].Value)
	assert.False(t, values["password"].Generated)
	assert.Equal(t, int64(16), values["workers"].Value)
	assert.Equal(t, "16", values["workers"].Text())
	assert.Equal(t, unit.DataSize(1<<30), values["memoryLimit"].Value)
	assert.Equal(t, "1gb", values["memoryLimit"].Text())
	assert.Equal(t, true, values["debug"].Value)
}

func TestResolveParamsRefuses(t *testing.T) {
	cases := map[string]struct {
		input map[string]any
		want  string
	}{
		"unknown parameter": {map[string]any{"dataVolume": "v", "surprise": 1}, "surprise"},
		"missing volume":    {map[string]any{}, "dataVolume: a value is required"},
		"pattern":           {map[string]any{"dataVolume": "v", "dbName": "Shop"}, "dbName: must match"},
		"too long":          {map[string]any{"dataVolume": "v", "dbName": "abcdefghi"}, "dbName: must be at most 8"},
		"not a number":      {map[string]any{"dataVolume": "v", "workers": "many"}, "workers: must be a whole number"},
		"fraction":          {map[string]any{"dataVolume": "v", "workers": 1.5}, "workers: must be a whole number"},
		"above max":         {map[string]any{"dataVolume": "v", "workers": 65}, "workers: must be at most 64"},
		"below min size":    {map[string]any{"dataVolume": "v", "memoryLimit": "64MB"}, "memoryLimit: must be at least"},
		"not a size":        {map[string]any{"dataVolume": "v", "memoryLimit": "lots"}, "memoryLimit: must be a size"},
		"not a bool":        {map[string]any{"dataVolume": "v", "debug": "maybe"}, "debug: must be true or false"},
		"unknown option":    {map[string]any{"dataVolume": "v", "mode": "turbo"}, "mode: must be one of"},
		"secret not text":   {map[string]any{"dataVolume": "v", "password": 42}, "password: must be text"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ResolveParams(testParams(), tc.input)
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
			assert.Contains(t, paramDetail(t, err), tc.want)
		})
	}
}

func TestResolveParamsGeneratesHex(t *testing.T) {
	defs := []*templatemodel.Parameter{{Name: "key", Type: templatemodel.ParamTypeSecret,
		Generate: &templatemodel.Generate{Length: 32, Charset: templatemodel.CharsetHex}}}
	values, err := ResolveParams(defs, nil)
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), values["key"].Text())
}

func TestValidateDefaults(t *testing.T) {
	assert.NoError(t, ValidateDefaults(testParams()))

	broken := testParams()
	broken[4].Default = "64MB" // below its own min
	err := ValidateDefaults(broken)
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
	assert.Contains(t, paramDetail(t, err), "memoryLimit")
}

func TestResolveParamsAppKey(t *testing.T) {
	defs := []*templatemodel.Parameter{
		{Name: "primaryApp", Title: "Primary", Type: templatemodel.ParamTypeApp, Engine: "postgres"},
	}

	values, err := ResolveParams(defs, map[string]any{"primaryApp": "  shop-db  "})
	assert.NoError(t, err)
	assert.Equal(t, "shop-db", values["primaryApp"].Value)

	// Apps created before keys became host names keep their underscores.
	values, err = ResolveParams(defs, map[string]any{"primaryApp": "shop_db"})
	assert.NoError(t, err)
	assert.Equal(t, "shop_db", values["primaryApp"].Value)

	for name, given := range map[string]any{
		"a leading dash": "-shop-db",
		"upper case":     "ShopDB",
		"a dot":          "shop.db",
		"not text":       42,
		"empty":          "",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ResolveParams(defs, map[string]any{"primaryApp": given})
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
			assert.Contains(t, paramDetail(t, err), "primaryApp")
		})
	}
}
