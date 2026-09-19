package templatemodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func withDependency(t *testing.T) *Template {
	t.Helper()
	tmpl := mustDecode(t)
	tmpl.Dependencies = []*Dependency{{
		Name: "db", Title: "Database", Template: "mysql", Version: "8.4",
		Params: map[string]any{"dbName": "wordpress"},
	}}
	return tmpl
}

func TestValidateAcceptsADependency(t *testing.T) {
	assert.NoError(t, withDependency(t).Validate("demo"))
}

func TestValidateRefusesABadDependency(t *testing.T) {
	cases := map[string]struct {
		mutate func(tmpl *Template)
		want   string
	}{
		"too many": {func(tm *Template) {
			for _, name := range []string{"cache", "search", "queue", "convert", "extract"} {
				tm.Dependencies = append(tm.Dependencies, &Dependency{Name: name, Title: name, Template: "redis"})
			}
		}, "at most 5"},
		"uppercase name": {func(tm *Template) { tm.Dependencies[0].Name = "DB" }, "lowercase letters and digits"},
		"dashed name":    {func(tm *Template) { tm.Dependencies[0].Name = "my-db" }, "lowercase letters and digits"},
		"no title":       {func(tm *Template) { tm.Dependencies[0].Title = "" }, "title is required"},
		"bad template":   {func(tm *Template) { tm.Dependencies[0].Template = "My SQL" }, "is not a template name"},
		"itself":         {func(tm *Template) { tm.Dependencies[0].Template = "demo" }, "cannot depend on itself"},
		"bad version":    {func(tm *Template) { tm.Dependencies[0].Version = "8.4 lts" }, `version "8.4 lts" is invalid`},
		"bad param name": {func(tm *Template) { tm.Dependencies[0].Params["db name"] = "x" }, "is not a parameter name"},
		"declared twice": {func(tm *Template) {
			tm.Dependencies = append(tm.Dependencies, &Dependency{Name: "db", Title: "Again", Template: "postgres"})
		}, "declared twice"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tmpl := withDependency(t)
			tc.mutate(tmpl)
			err := tmpl.Validate("demo")
			assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
			assert.Contains(t, errorDetail(t, err), tc.want)
		})
	}
}

func TestAskedParams(t *testing.T) {
	database := &Template{Parameters: []*Parameter{
		{Name: "dbName", Type: ParamTypeString, Default: "app"},
		{Name: "username", Type: ParamTypeString},
		{Name: "password", Type: ParamTypeSecret, Generate: &Generate{Length: 32}},
		{Name: "note", Type: ParamTypeString, Optional: true},
		{Name: "blank", Type: ParamTypeString, Default: ""},
		{Name: "dataVolume", Type: ParamTypeVolume},
	}}
	dep := &Dependency{Name: "db", Params: map[string]any{"username": "wordpress"}}

	var names []string
	for _, param := range dep.AskedParams(database) {
		names = append(names, param.Name)
	}

	assert.Equal(t, []string{"blank", "dataVolume"}, names,
		"fixed, defaulted, optional and generated parameters are not asked; an empty default is no default")
}

func TestDependencyAppName(t *testing.T) {
	assert.Equal(t, "blog-db", DependencyAppName("blog", "db"))
}
