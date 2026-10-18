package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// A template that needs a feature the license does not grant creates nothing,
// and is refused before anything is provisioned.
func TestATemplateNeedingAFeatureTheLicenseLacksCreatesNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.resp.Template.Metadata.Requires.Features = []string{"managed-databases"}
	fakes.licenses.featuresErr = hperrors.Wrap(hperrors.ErrPlanFeatureRequired)

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrPlanFeatureRequired)
	assert.False(t, fakes.provision.called)
	assert.Equal(t, [][]string{{"managed-databases"}}, fakes.licenses.features)
}

// What a dependency needs is needed too, once: the person creating the app is
// creating the dependency as well.
func TestADependencysFeaturesAreChecked(t *testing.T) {
	uc, fakes := newCreateTest(t)
	main := fakes.templates.resp
	main.Template.Metadata.Requires.Features = []string{"managed-databases"}
	depTemplate := *main.Template
	depTemplate.Metadata.Requires = templatemodel.Requires{
		VersionCode: "v000002", Features: []string{"managed-databases", "backups-plus"},
	}
	main.Dependencies = []*apptemplateservice.RenderedDependency{{
		Name: "db", AppName: "main-db-db",
		Render: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{Entry: main.Entry, Template: &depTemplate},
			Result:       main.Result,
		},
	}}
	fakes.licenses.featuresErr = hperrors.Wrap(hperrors.ErrPlanFeatureRequired)

	_, err := uc.CreateAppFromTemplate(context.Background(), testAuth(), testCreateReq())

	assert.ErrorIs(t, err, hperrors.ErrPlanFeatureRequired)
	assert.Equal(t, [][]string{{"backups-plus", "managed-databases"}}, fakes.licenses.features)
}

// A template that needs nothing asks the license nothing.
func TestATemplateNeedingNoFeatureAsksNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)

	assert.NoError(t, uc.checkFeatures(context.Background(), planApps(testCreateReq(), fakes.templates.resp)))
	assert.Empty(t, fakes.licenses.features)
}
