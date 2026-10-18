package apptemplatedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type grants map[string]bool

func (g grants) Grants(feature string) bool { return g[feature] }

func summaryOf(entry *templatemodel.IndexEntry, granted FeatureGrants) *AppTemplateSummaryResp {
	return TransformAppTemplateSummaries([]*templatemodel.IndexEntry{entry}, "v000002", granted)[0]
}

// The store says what a template needs, and which of it the license in force
// does not grant.
func TestTheSummarySaysWhatFeaturesAreMissing(t *testing.T) {
	useBasePath(t)
	entry := testEntry("v000002")
	entry.Managed = "postgres"
	entry.Requires.Features = []string{"managed-databases", "x"}

	granted := summaryOf(entry, grants{"managed-databases": true})
	assert.Equal(t, "postgres", granted.Managed)
	assert.Equal(t, []string{"managed-databases", "x"}, granted.RequiresFeatures)
	assert.Equal(t, []string{"x"}, granted.MissingFeatures)

	assert.Equal(t, []string{"managed-databases", "x"}, summaryOf(entry, nil).MissingFeatures,
		"no license grants nothing")

	plain := summaryOf(testEntry("v000001"), nil)
	assert.Equal(t, []string{}, plain.RequiresFeatures, "a list, never null")
	assert.Equal(t, []string{}, plain.MissingFeatures)
	assert.Empty(t, plain.Managed)
}
