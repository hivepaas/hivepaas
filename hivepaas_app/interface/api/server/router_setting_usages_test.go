package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// notSettings are groups deleted by id that are not settings: nothing links to
// a snapshot or a data file, and their deletes never answer ERR_SETTING_IN_USE.
var notSettings = []string{"/backup-snapshots/:itemID", "/data-files/:itemID"}

// A setting refused deletion because something uses it has what uses it listed
// under the path the delete was sent to - the dashboard asks there - whether it
// is the installation's, a project's, an environment's or an app's.
func TestEverySettingDeletedByIDListsItsUsages(t *testing.T) {
	routes := map[string]bool{}
	for _, route := range allRoutes(t) {
		routes[route.Method+" "+route.Path] = true
	}

	checked := 0
	for route := range routes {
		path, isDelete := strings.CutPrefix(route, "DELETE ")
		if !isDelete || !strings.HasSuffix(path, "/:itemID") ||
			!(strings.HasPrefix(path, "/api/settings/") || strings.HasPrefix(path, "/api/projects/")) {
			continue
		}
		if strings.HasSuffix(path, notSettings[0]) || strings.HasSuffix(path, notSettings[1]) {
			continue
		}
		checked++
		assert.True(t, routes["GET "+path+"/usages"], "%s has no usages route", path)
	}
	assert.Greater(t, checked, 50)
}
