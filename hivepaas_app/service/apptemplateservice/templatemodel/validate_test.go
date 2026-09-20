package templatemodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func mustDecode(t *testing.T) *Template {
	t.Helper()
	tmpl, err := DecodeTemplate([]byte(validTemplateYAML))
	assert.NoError(t, err)
	return tmpl
}

func TestValidateAcceptsAValidTemplate(t *testing.T) {
	assert.NoError(t, mustDecode(t).Validate("demo"))
}

func TestValidateReportsEveryProblem(t *testing.T) {
	cases := map[string]struct {
		mutate func(tmpl *Template)
		want   string
	}{
		"wrong kind":           {func(tm *Template) { tm.Kind = "Spec" }, `kind must be "AppTemplate"`},
		"file name mismatch":   {func(tm *Template) { tm.Metadata.Name = "other" }, "must equal the file name"},
		"uppercase name":       {func(tm *Template) { tm.Metadata.Name = "Demo" }, "lowercase"},
		"long tagline":         {func(tm *Template) { tm.Metadata.Tagline = strings.Repeat("x", 81) }, "tagline"},
		"no categories":        {func(tm *Template) { tm.Metadata.Categories = nil }, "categories must have 1 to 3"},
		"flat category":        {func(tm *Template) { tm.Metadata.Categories = []string{"sql"} }, "parent/child"},
		"too many tags":        {func(tm *Template) { tm.Metadata.Tags = strings.Split("a,b,c,d,e,f,g,h,i", ",") }, "tags"},
		"jpeg icon":            {func(tm *Template) { tm.Metadata.Icon = "icons/demo.jpg" }, ".svg or .png"},
		"http link":            {func(tm *Template) { tm.Metadata.Links.Website = "http://example.com" }, "https"},
		"bad version code":     {func(tm *Template) { tm.Metadata.Requires.VersionCode = "1" }, "versionCode"},
		"secret default":       {func(tm *Template) { tm.Parameters[0].Default = "hunter2" }, "a secret has no default"},
		"short generate":       {func(tm *Template) { tm.Parameters[0].Generate.Length = 4 }, "generate.length"},
		"unknown charset":      {func(tm *Template) { tm.Parameters[0].Generate.Charset = "emoji" }, "charset"},
		"bad size bound":       {func(tm *Template) { tm.Parameters[1].Min = "lots" }, "min and max must be sizes"},
		"pattern on size":      {func(tm *Template) { tm.Parameters[1].Pattern = "^a$" }, "pattern does not apply"},
		"volume default":       {func(tm *Template) { tm.Parameters[2].Default = "vol" }, "a volume has no default"},
		"unknown type":         {func(tm *Template) { tm.Parameters[2].Type = "ssl-cert" }, "type \"ssl-cert\""},
		"duplicate parameter":  {func(tm *Template) { tm.Parameters[2].Name = "password" }, "declared twice"},
		"two default variants": {func(tm *Template) { tm.Variants[1].Default = true }, "exactly one variant"},
		"two default versions": {func(tm *Template) { tm.Versions[1].Default = true }, "exactly one version"},
		"deprecated default":   {func(tm *Template) { tm.Versions[0].Deprecated = true }, "cannot be deprecated"},
		"image with variants":  {func(tm *Template) { tm.Versions[0].Image = "demo:2.1.0" }, "set images and not image"},
		"undeclared variant": {func(tm *Template) {
			tm.Versions[0].Images["musl"] = "demo:2.1.0-musl"
		}, `"musl" is not declared`},
		"floating tag":   {func(tm *Template) { tm.Versions[0].Images["alpine"] = "demo:2-alpine" }, "exact release"},
		"latest tag":     {func(tm *Template) { tm.Versions[0].Images["debian"] = "demo:latest" }, "exact release"},
		"empty override": {func(tm *Template) { tm.Versions[0].Override = &Override{} }, "override.app is empty"},
		"no app":         {func(tm *Template) { tm.App = nil }, "app is required"},
		"no versions":    {func(tm *Template) { tm.Versions = nil }, "versions must not be empty"},
		"select without option": {func(tm *Template) {
			tm.Parameters = append(tm.Parameters, &Parameter{Name: "mode", Title: "Mode", Type: ParamTypeSelect})
		}, "options"},
		"app default": {func(tm *Template) {
			tm.Parameters = append(tm.Parameters, &Parameter{Name: "primary", Title: "Primary",
				Type: ParamTypeApp, Default: "shop_db"})
		}, "an app parameter has no default"},
		"app engine shape": {func(tm *Template) {
			tm.Parameters = append(tm.Parameters, &Parameter{Name: "primary", Title: "Primary",
				Type: ParamTypeApp, Engine: "Postgres!"})
		}, "engine \"Postgres!\""},
		"engine on another type": {func(tm *Template) {
			tm.Parameters = append(tm.Parameters, &Parameter{Name: "region", Title: "Region",
				Type: ParamTypeString, Engine: "postgres"})
		}, "engine does not apply to type string"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tmpl := mustDecode(t)
			tc.mutate(tmpl)
			err := tmpl.Validate("demo")
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
			assert.Contains(t, errorDetail(t, err), tc.want)
		})
	}
}

func TestValidateWithoutVariantsWantsImage(t *testing.T) {
	tmpl := mustDecode(t)
	tmpl.Variants = nil
	err := tmpl.Validate("demo")
	assert.Contains(t, errorDetail(t, err), "without variants, set image and not images")

	for _, version := range tmpl.Versions {
		version.Image, version.Images = version.Images["alpine"], nil
	}
	assert.NoError(t, tmpl.Validate("demo"))
}

func TestIsPinnedImage(t *testing.T) {
	for image, pinned := range map[string]bool{
		"postgres:17.6-alpine3.22":                   true,
		"postgres:17.6":                              true,
		"mariadb:11.8.3-noble":                       true,
		"minio/minio:RELEASE.2025-09-07T16-13-09Z":   true,
		"registry.example:5000/app:1.2.3":            true,
		"postgres@sha256:" + strings.Repeat("a", 64): true,
		// A word in front of the version is the only shape some repositories
		// publish: fireflyiii/core has version-6.7.2 and nothing plainer.
		"fireflyiii/core:version-6.7.2": true,
		"example/app:release-2.4":       true,
		"example/app:v-1.2.3":           true,
		// The base image in the suffix carries a version of its own, and a tag is
		// only pinned when the software's own version part is.
		"postgres:18-alpine3.24": false,
		// A word and then something that is not a version is still a moving tag,
		// whatever the suffix carries.
		"example/app:stable-alpine3.24": false,
		"example/app:version-6":         false,
		"example/app:edge":              false,
		"postgres:19beta1-alpine3.24":   false,
		"postgres:17":                   false,
		"postgres:17-alpine":            false,
		"postgres:latest":               false,
		"postgres":                      false,
		"registry.example:5000/app":     false,
	} {
		assert.Equal(t, pinned, IsPinnedImage(image), image)
	}
}

func TestParameterBounds(t *testing.T) {
	sizeParam := &Parameter{Type: ParamTypeSize, Min: "128MB", Max: "1GB"}
	minSize, maxSize := sizeParam.SizeBounds()
	assert.Equal(t, int64(128<<20), minSize.Bytes())
	assert.Equal(t, int64(1<<30), maxSize.Bytes())

	intParam := &Parameter{Type: ParamTypeInt, Min: 1, Max: float64(10)}
	minInt, maxInt := intParam.IntBounds()
	assert.Equal(t, int64(1), *minInt)
	assert.Equal(t, int64(10), *maxInt)

	unbounded := &Parameter{Type: ParamTypeInt}
	minInt, maxInt = unbounded.IntBounds()
	assert.Nil(t, minInt)
	assert.Nil(t, maxInt)
}

func TestValidateAcceptsAnAppParameter(t *testing.T) {
	tmpl := mustDecode(t)
	tmpl.Parameters = append(tmpl.Parameters, &Parameter{
		Name: "primaryApp", Title: "Primary app key", Type: ParamTypeApp, Engine: "postgres",
	})
	assert.NoError(t, tmpl.Validate("demo"))
}
