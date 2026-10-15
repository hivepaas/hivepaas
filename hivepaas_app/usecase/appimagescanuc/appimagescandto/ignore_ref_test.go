package appimagescandto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
)

// An ignore names what it is set on: a preview's own and its parent app's are
// both of the app scope, and only this says where to remove each.
func TestAnIgnoreNamesWhatItIsSetOn(t *testing.T) {
	ignore := &imagescanservice.Ignore{
		Setting: &entity.Setting{ID: "ig1", Scope: base.ObjectScopeApp, ObjectID: "parent-app", Inheritable: true},
		Data:    &entity.VulnerabilityIgnore{VulnID: "CVE-2025-1", Reason: "not reachable"},
	}

	resp := TransformIgnoreRef(ignore)

	assert.Equal(t, "ig1", resp.ID)
	assert.Equal(t, base.ObjectScopeApp, resp.Scope)
	assert.Equal(t, "parent-app", resp.ObjectID)
}
