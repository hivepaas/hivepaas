package licenseserviceimpl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
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

// A development installation is given the features HP_DEV_MODE_LICENSE_FEATURES
// names, on Bee or on a key without them, and the template store sees them too.
func TestDevModeGrantsItsFeatures(t *testing.T) {
	h := newHarness(t)
	h.config.Env = config.EnvDev
	h.config.DevMode = config.DevMode{Enabled: true, LicenseFeatures: "managed-databases"}

	assert.NoError(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}), "on Bee")
	state, err := h.Current(context.Background())
	assert.NoError(t, err)
	assert.True(t, state.Grants("managed-databases"))

	h.storeKey(t, payload(func(p *license.Payload) { p.Features = nil }))
	assert.NoError(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}), "with a key")
}

// Where dev mode is off, the setting grants nothing.
func TestDevFeaturesAreIgnoredOutsideDevMode(t *testing.T) {
	h := newHarness(t)
	h.config.DevMode = config.DevMode{LicenseFeatures: "managed-databases"}

	assert.ErrorIs(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}),
		hperrors.ErrPlanFeatureRequired)
}

// A beta installation has managed databases on Bee and on a key without them:
// they are free while HivePaaS is in beta.
func TestABetaInstallationHasTheBetaFeatures(t *testing.T) {
	h := newHarness(t)
	h.config.Env = config.EnvBeta

	assert.NoError(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}), "on Bee")
	state, err := h.Current(context.Background())
	assert.NoError(t, err)
	assert.True(t, state.Grants("managed-databases"), "the template store sees it too")

	h.storeKey(t, payload(func(p *license.Payload) { p.Features = nil }))
	assert.NoError(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}), "with a key")
}

// A stable installation has them only from its key.
func TestAStableInstallationHasNoBetaFeatures(t *testing.T) {
	h := newHarness(t)

	assert.ErrorIs(t, h.CheckFeatures(context.Background(), []string{"managed-databases"}),
		hperrors.ErrPlanFeatureRequired)
}
