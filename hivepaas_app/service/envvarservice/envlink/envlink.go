// Package envlink suggests the variables that link one app to another of its
// env: references to what the target shares, put together the way a client of
// its engine wants them. It holds no values: every suggestion is a reference,
// resolved when the app is built.
package envlink

import (
	"regexp"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// Target is an app another app of its env may link to.
type Target struct {
	ID       string
	Key      string
	Name     string
	Category base.AppCategory
	Engine   string
}

// Facts are what the suggestions depend on besides the target's kind.
type Facts struct {
	HasPort     bool
	HasPassword bool
	// SharedVars are the keys of the variables the target's own settings share,
	// not its system variables.
	SharedVars []string
}

// Var is one suggested variable.
type Var struct {
	Key         string
	Value       string
	Description string
}

// Group is suggested variables that go together.
type Group struct {
	ID          string
	Title       string
	Description string
	Recommended bool
	Warnings    []string
	Vars        []*Var
}

// Targets are the apps of self's env another may link to: not self, not a
// preview, ordered by name.
func Targets(self *entity.App, apps []*entity.App, kinds map[string]*entity.AppKindSettings) []*Target {
	targets := make([]*Target, 0, len(apps))
	for _, app := range apps {
		if app.ID == self.ID || app.ParentID != "" || app.ProjectEnvID != self.ProjectEnvID {
			continue
		}
		target := &Target{ID: app.ID, Key: app.Key, Name: app.Name}
		if kind := kinds[app.ID]; kind != nil {
			target.Category, target.Engine = kind.Category, kind.Engine
		}
		targets = append(targets, target)
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].Name < targets[j].Name })
	return targets
}

// FactsOf reads the facts from the target's shared environment.
func FactsOf(shared []*envvarservice.EnvVar) *Facts {
	facts := &Facts{}
	for _, env := range shared {
		switch {
		case env.IsSystem && env.Key == base.AppSystemEnvVarPort:
			facts.HasPort = env.Value != ""
		case env.IsSystem && env.Key == base.AppSystemEnvVarPassword:
			facts.HasPassword = env.Value != ""
		case !env.IsSystem:
			facts.SharedVars = append(facts.SharedVars, env.Key)
		}
	}
	return facts
}

var notEnvKeyChars = regexp.MustCompile(`[^A-Z0-9_]+`)

// envKeyOf is an app key made a variable name: upper case, anything else an
// underscore, and never starting with a digit.
func envKeyOf(key string) string {
	name := notEnvKeyChars.ReplaceAllString(strings.ToUpper(key), "_")
	if name == "" || (name[0] >= '0' && name[0] <= '9') {
		name = "APP_" + name
	}
	return name
}

var reReference = regexp.MustCompile(`\$\{([a-zA-Z0-9_-]+)\.([A-Za-z_][A-Za-z0-9_]*)\}`)

// referencedNames are the variables of app a value refers to.
func referencedNames(value, app string) []string {
	var names []string
	for _, m := range reReference.FindAllStringSubmatch(value, -1) {
		if m[1] == app {
			names = append(names, m[2])
		}
	}
	return names
}
