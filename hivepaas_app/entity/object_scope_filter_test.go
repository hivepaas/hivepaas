package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A list's filters narrow what the scope it was reached through shows, and
// never widen it: through a project, another project's tasks or audit log were
// listed by naming it in a filter.
func TestObjectScopeFilteredBy(t *testing.T) {
	project := NewObjectScopeProject("p1")
	keptToDev := NewObjectScopeProject("p1")
	keptToDev.ProjectEnvIDs = []string{"p1:dev"}
	dev := NewObjectScopeProjectEnv("p1", "dev")
	app := NewObjectScopeApp("a1", "", "p1", "p1:dev")
	prodApp := NewObjectScopeApp("a2", "", "p1", "p1:prod")
	otherApp := NewObjectScopeApp("a3", "", "p2", "p2:dev")

	tests := []struct {
		name   string
		scope  *ObjectScope
		filter *ObjectScope
		want   *ObjectScope
	}{
		{name: "no filter", scope: project, filter: nil, want: project},
		{name: "everything, to another project", scope: NewObjectScopeGlobal(), filter: NewObjectScopeProject("p2"),
			want: NewObjectScopeProject("p2")},
		{name: "a project, to one of its envs", scope: project, filter: dev, want: dev},
		{name: "a project, to one of its apps", scope: project, filter: prodApp, want: prodApp},
		{name: "a project, to itself", scope: project, filter: NewObjectScopeProject("p1"), want: project},
		{name: "a project, to another project", scope: project, filter: NewObjectScopeProject("p2"), want: nil},
		{name: "a project, to another project's env", scope: project, filter: NewObjectScopeProjectEnv("p2", "dev"),
			want: nil},
		{name: "a project, to another project's app", scope: project, filter: otherApp, want: nil},
		{name: "a project kept to an env, to that env", scope: keptToDev, filter: dev, want: dev},
		{name: "a project kept to an env, to an app of it", scope: keptToDev, filter: app, want: app},
		{name: "a project kept to an env, to another env", scope: keptToDev,
			filter: NewObjectScopeProjectEnv("p1", "prod"), want: nil},
		{name: "a project kept to an env, to an app of another", scope: keptToDev, filter: prodApp, want: nil},
		{name: "a project kept to an env, to itself", scope: keptToDev, filter: NewObjectScopeProject("p1"),
			want: keptToDev},
		{name: "an env, to an app of it", scope: dev, filter: app, want: app},
		{name: "an env, to an app of another env", scope: dev, filter: prodApp, want: nil},
		{name: "an env, to its project", scope: dev, filter: NewObjectScopeProject("p1"), want: dev},
		{name: "an env, to another project", scope: dev, filter: NewObjectScopeProject("p2"), want: nil},
		{name: "an app, to its env", scope: app, filter: dev, want: app},
		{name: "an app, to another app", scope: app, filter: prodApp, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.scope.FilteredBy(tt.filter))
		})
	}
}
