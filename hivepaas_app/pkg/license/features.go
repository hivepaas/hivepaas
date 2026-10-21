package license

import (
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// FeaturesOf is the features a key puts in force: those of its features this
// binary knows, in base.AllLicenseFeatures order. A key grants nothing it does
// not list.
func FeaturesOf(p *Payload) []base.LicenseFeature {
	var features []base.LicenseFeature
	for _, feature := range base.AllLicenseFeatures {
		if slices.Contains(p.Features, string(feature)) {
			features = append(features, feature)
		}
	}
	return features
}

// Grants reports whether the license in force grants a feature, named as a
// template names it. A name this binary does not know is never granted.
func (s *State) Grants(feature string) bool {
	return s != nil && slices.Contains(s.Features, base.LicenseFeature(feature))
}

// WithFeatures puts these features in force besides the state's own: those of
// them this binary knows, in base.AllLicenseFeatures order. They are a
// development installation's (HP_DEV_MODE_LICENSE_FEATURES) or a beta one's
// (base.BetaLicenseFeatures), never a key's.
func (s *State) WithFeatures(features []string) *State {
	if s == nil || len(features) == 0 {
		return s
	}
	var inForce []base.LicenseFeature
	for _, feature := range base.AllLicenseFeatures {
		if slices.Contains(s.Features, feature) || slices.Contains(features, string(feature)) {
			inForce = append(inForce, feature)
		}
	}
	s.Features = inForce
	return s
}

// featureNameMsgIDs name each feature in ERR_PLAN_FEATURE_REQUIRED. A test holds
// it to base.AllLicenseFeatures.
var featureNameMsgIDs = map[base.LicenseFeature]string{
	base.LicenseFeatureManagedDatabases: "ERR_PLAN_FEATURE_MANAGED_DATABASES",
}

// FeatureNameMsgID is the message id of what ERR_PLAN_FEATURE_REQUIRED calls a
// feature, and "" for one this binary does not know, which is named as written.
func FeatureNameMsgID(feature base.LicenseFeature) string {
	return featureNameMsgIDs[feature]
}
