package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// Every type a document can hold is imported by a policy or skipped with a
// reason: none is written because nobody thought about it.
func TestEveryBlockTypeHasAnImportPolicy(t *testing.T) {
	for _, typ := range specmodel.BlockTypes() {
		_, listed := importPolicies[typ]
		assert.True(t, listed, "%s has no import policy and no reason to skip it", typ)
	}
}

// A setting whose writing does more than its row is not written, and the plan
// says why; the rest of its scope is.
func TestPlanSkipsASettingOfATypeImportDoesNotWrite(t *testing.T) {
	svc, bundle := planFixture(t)
	settings := bundle.Projects["project_a"].Settings
	settings["periodicJobs"] = map[string]any{"nightly": map[string]any{}}
	settings["basicAuths"] = map[string]any{"admin": map[string]any{"username": "admin"}}

	project := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), "projects/project_a/settings")

	assert.Equal(t, []string{"basicAuths/admin"}, project.Changes)
	if issues := issuesOf(project, specmodel.CodeTypeNotImportable); assert.Len(t, issues, 1) {
		assert.Equal(t, specmodel.SeveritySkipped, issues[0].Severity)
		assert.Equal(t, "periodicJobs/nightly", issues[0].Detail["setting"])
		assert.Equal(t, reasonSchedulesTasks, issues[0].Detail["reason"])
	}
}
