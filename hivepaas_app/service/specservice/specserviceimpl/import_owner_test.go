package specserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const projectPath = "projects/project_a"

func setOwner(bundle *specmodel.ImportBundle, id, email string) {
	bundle.Projects["project_a"].Owner = &specmodel.ProjectOwner{ID: id, Email: email}
}

// On another installation the owner has another id; their email finds them,
// and the project's owner is who it already was.
func TestPlanFindsTheOwnerByEmailWhenTheIDIsAnotherInstallations(t *testing.T) {
	svc, bundle := planFixture(t)
	setOwner(bundle, "u_elsewhere", "OWNER@example.com")

	project := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), projectPath)

	assert.Equal(t, specmodel.ActionUnchanged, project.Action)
	assert.Empty(t, project.Notes)
}

func TestPlanChangesTheOwnerOnlyWhereTheOperatorMay(t *testing.T) {
	for allowed, want := range map[bool][]string{true: {"owner"}, false: nil} {
		svc, bundle := planFixture(t)
		setOwner(bundle, "u_elsewhere", "other@example.com")
		var askedFor string

		project := node(t, planWith(t, svc, bundle, &specservice.ValidateImportReq{
			MayChangeOwner: func(_ context.Context, project *entity.Project) (bool, error) {
				askedFor = project.ID
				return allowed, nil
			},
		}), projectPath)

		assert.Equal(t, "p1", askedFor)
		assert.Equal(t, want, project.Changes)
		issues := issuesOf(project, specmodel.CodeOwnerNotPermitted)
		if allowed {
			assert.Empty(t, issues)
			continue
		}
		if assert.Len(t, issues, 1) {
			assert.Equal(t, specmodel.SeverityFixable, issues[0].Severity)
			assert.Equal(t, map[string]any{"owner": "other@example.com"}, issues[0].Detail)
		}
	}
}

// Nobody here is the owner - a disabled user does not count - so an existing
// project keeps its owner, and says so.
func TestPlanKeepsTheOwnerOfAnExistingProjectWhenNobodyMatches(t *testing.T) {
	svc, bundle := planFixture(t)
	setOwner(bundle, "u2", "gone@example.com")

	project := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), projectPath)

	assert.Equal(t, specmodel.ActionUnchanged, project.Action)
	if notes := notesOf(project, specmodel.CodeOwnerNotFound); assert.Len(t, notes, 1) {
		assert.Equal(t, map[string]any{"owner": "gone@example.com"}, notes[0].Detail)
		assert.Equal(t, "the project keeps its owner", notes[0].Action)
	}
}

func TestPlanGivesANewProjectWhoseOwnerIsNobodyToTheOperator(t *testing.T) {
	svc, bundle := planFixture(t)
	other, err := readBundle(exportedBytes(t, specmodel.SecretsModeOmit, ""), "")
	assert.NoError(t, err)
	project := other.Projects["project_a"]
	project.ID, project.Name = "", "Project New"
	project.Owner = &specmodel.ProjectOwner{ID: "u_elsewhere", Email: "nobody@example.com"}
	bundle.Projects["project_new"] = project

	created := node(t, plan(t, svc, bundle, specmodel.ImportOptions{}), "projects/project_new")

	assert.Equal(t, specmodel.ActionCreate, created.Action)
	if notes := notesOf(created, specmodel.CodeOwnerNotFound); assert.Len(t, notes, 1) {
		assert.Equal(t, "the operator importing owns the project", notes[0].Action)
	}
}
