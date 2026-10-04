package specserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const newAppPath = "projects/blog/envs/prod/apps/db"

// newProjectBundle is a bundle built in memory, as a compose file is read: a
// project the installation does not have, with one env and an app mounting the
// project's default volume, which the bundle does not carry.
func newProjectBundle() *specmodel.ImportBundle {
	header := specmodel.NewDocHeader("project")
	return &specmodel.ImportBundle{
		Manifest: &specmodel.Manifest{APIVersion: specmodel.APIVersion, Kind: specmodel.KindSpec,
			Scope: "project", SecretsMode: specmodel.SecretsModePlaintext},
		Projects: map[string]*specmodel.ProjectDoc{"blog": {DocHeader: header, Project: "blog", Name: "Blog"}},
		Envs: map[string]map[string]*specmodel.EnvDoc{"blog": {"prod": {
			DocHeader: header, Project: "blog", Env: "prod", Name: "production",
			Apps: map[string]*specmodel.AppDoc{"db": {App: "db", Name: "db", Deployment: &specmodel.Deployment{
				Source: map[string]any{"activeMethod": "image", "imageSource": map[string]any{"image": "postgres:17"}},
				Storage: &specmodel.Storage{Mounts: map[string]specmodel.Mount{"/var/lib/postgresql/data": {
					Type: mount.TypeVolume, Source: "projects/blog/volumes/default",
					VolumeOptions: &specmodel.VolumeOptions{Subpath: "pgdata"},
				}}},
			}}},
		}}},
		Digest: "digest-1",
	}
}

func planBundleReq(issues map[string][]specmodel.Issue) *specservice.PlanBundleReq {
	return &specservice.PlanBundleReq{
		ValidateImportReq: specservice.ValidateImportReq{
			Scope: entity.NewObjectScopeGlobal(), Options: specmodel.ImportOptions{Existing: specmodel.ExistingUpdate},
		},
		Doc: newProjectBundle(), Issues: issues,
	}
}

// A new project's default volume is mounted before it exists: it is made with
// the project, so the plan finds nothing missing.
func TestPlanBundleMountsANewProjectsDefaultVolume(t *testing.T) {
	svc, _ := planFixture(t)
	plan, err := svc.PlanBundle(context.Background(), nil, planBundleReq(nil))
	assert.NoError(t, err)

	app := node(t, plan, newAppPath)
	assert.Equal(t, specmodel.ActionCreate, app.Action)
	assert.Empty(t, app.Issues)
	assert.Equal(t, specmodel.ActionCreate, node(t, plan, "projects/blog").Action)
}

// Applied, the mount reaches the volume the project was given.
func TestApplyBundleMountsTheVolumeTheProjectWasGiven(t *testing.T) {
	svc, _ := planFixture(t)
	req := planBundleReq(nil)
	plan, err := svc.PlanBundle(context.Background(), nil, req)
	assert.NoError(t, err)

	resp, err := svc.ApplyBundle(context.Background(), nil, &specservice.ApplyBundleReq{
		PlanBundleReq: *req, OperatorID: "u_operator", PlanHash: plan.PlanHash, AcceptIssues: true,
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, specmodel.OutcomeApplied, node(t, resp.Plan, newAppPath).Outcome)

	project := persisted(svc).UpsertingProjects[0]
	provision := svc.appProvisionService.(*fakeProvisionService)
	if assert.Len(t, provision.reqs, 1) {
		mounts := provision.specs[provision.reqs[0].AppID].TaskTemplate.ContainerSpec.Mounts
		if assert.Len(t, mounts, 1) {
			assert.Equal(t, "volume_"+project.ID, mounts[0].Source)
		}
	}
}

// The reader's issues are the plan's: a note among the notes, a skipped one
// skipping its node, and each counted in the hash.
func TestPlanBundleTakesTheIssuesOfItsReading(t *testing.T) {
	svc, _ := planFixture(t)
	plain, err := svc.PlanBundle(context.Background(), nil, planBundleReq(nil))
	assert.NoError(t, err)

	plan, err := svc.PlanBundle(context.Background(), nil, planBundleReq(map[string][]specmodel.Issue{
		newAppPath: {
			{Code: "COMPOSE_ALIAS_ADDED", Path: newAppPath},
			{Severity: specmodel.SeverityWarning, Code: "COMPOSE_NOT_SUPPORTED", Path: newAppPath},
		},
	}))
	assert.NoError(t, err)
	app := node(t, plan, newAppPath)
	assert.Len(t, app.Notes, 1)
	assert.Len(t, app.Issues, 1)
	assert.NotEqual(t, plain.PlanHash, plan.PlanHash, "an issue is part of what was agreed to")

	plan, err = svc.PlanBundle(context.Background(), nil, planBundleReq(map[string][]specmodel.Issue{
		newAppPath: {{Severity: specmodel.SeveritySkipped, Code: "COMPOSE_NO_IMAGE", Path: newAppPath}},
	}))
	assert.NoError(t, err)
	assert.Equal(t, specmodel.ActionSkip, node(t, plan, newAppPath).Action)

	_, err = svc.PlanBundle(context.Background(), nil, planBundleReq(map[string][]specmodel.Issue{
		"projects/blog/envs/prod/apps/nothing": {{Code: "COMPOSE_ALIAS_ADDED"}},
	}))
	assert.Error(t, err, "an issue for no node is the reader's mistake")
}
