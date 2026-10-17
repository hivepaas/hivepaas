package licenseserviceimpl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// On Bee a feature is refused, with the feature and the plan on the wire for the
// dashboard's upgrade prompt.
func TestAFeatureBeeLacksIsRefused(t *testing.T) {
	h := newHarness(t)

	err := h.CheckFeatures(context.Background(), []string{"managed-databases"})

	assert.ErrorIs(t, err, hperrors.ErrPlanFeatureRequired)
	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, http.StatusPreconditionFailed, info.Status)
	assert.Equal(t, "ERR_PLAN_FEATURE_REQUIRED", info.Code)
	assert.Equal(t, map[string]any{"Feature": "managed-databases", "Plan": "bee"}, info.Params)
	assert.Equal(t, "The Bee plan does not include managed databases. Upgrade to use it.", info.Detail)
}

func TestAKeysFeaturesAreGranted(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload(func(p *license.Payload) { p.Features = []string{"managed-databases"} }))

	assert.NoError(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}))
}

// A feature this HivePaaS does not know cannot be granted, and is named as it is
// written.
func TestAnUnknownFeatureIsRefusedByItsName(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload(func(p *license.Payload) { p.Features = []string{"managed-databases"} }))

	err := h.CheckFeatures(context.Background(), []string{"managed-databases", "quantum"})

	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, map[string]any{"Feature": "quantum", "Plan": "hive"}, info.Params)
	assert.Equal(t, "The Hive plan does not include quantum. Upgrade to use it.", info.Detail)
}

// Needing nothing reads nothing.
func TestNeedingNoFeatureIsNotChecked(t *testing.T) {
	h := newHarness(t)

	assert.NoError(t, h.CheckFeatures(context.Background(), nil))
	assert.Zero(t, h.settings.reads)
}
