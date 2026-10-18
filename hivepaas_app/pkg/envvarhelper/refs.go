package envvarhelper

import (
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// AppRefPrefix is how a variable's value names another app's variables:
// ${<key>.VAR}. An app's key is matched whole, as references are resolved: the
// dot follows it.
func AppRefPrefix(appKey string) string {
	return "${" + appKey + "."
}

// RefersTo says whether a variable that is not literal names the app's
// variables.
func RefersTo(env *entity.EnvVar, appKey string) bool {
	return !env.IsLiteral && strings.Contains(env.Value, AppRefPrefix(appKey))
}

// RefReplacer rewrites references to apps' variables from one key to another:
// ${old.VAR} to ${new.VAR}, by the keys given, old to new.
func RefReplacer(keys map[string]string) *strings.Replacer {
	pairs := make([]string, 0, len(keys)*2) //nolint:mnd
	for _, old := range slices.Sorted(maps.Keys(keys)) {
		pairs = append(pairs, AppRefPrefix(old), AppRefPrefix(keys[old]))
	}
	return strings.NewReplacer(pairs...)
}

// RewriteRefs rewrites, in place, the references the replacer names in the
// variables that are not literal, and returns those it changed.
func RewriteRefs(vars []*entity.EnvVar, replacer *strings.Replacer) []*entity.EnvVar {
	var changed []*entity.EnvVar
	for _, env := range vars {
		if env.IsLiteral {
			continue
		}
		if value := replacer.Replace(env.Value); value != env.Value {
			env.Value = value
			changed = append(changed, env)
		}
	}
	return changed
}
