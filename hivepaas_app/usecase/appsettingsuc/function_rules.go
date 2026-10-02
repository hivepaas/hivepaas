package appsettingsuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A function is an app from its creation: its kind says so, and its other
// settings follow. These are the rules its settings' updates keep.

// checkDeploymentOfKind refuses deployment settings that do not suit the app: a
// function is deployed from its code, and only a function keeps a function's
// source.
func checkDeploymentOfKind(function bool, settings *entity.AppDeploymentSettings) error {
	isFunctionMethod := settings.ActiveMethod == base.DeploymentMethodFunction
	switch {
	case function && !isFunctionMethod:
		return hperrors.Wrap(hperrors.ErrDeploymentMethodFunctionRequired)
	case !function && (isFunctionMethod || settings.FunctionSource != nil):
		return hperrors.Wrap(hperrors.ErrDeploymentMethodFunctionUnallowed)
	}
	return nil
}

// checkKindCategoryChange refuses a kind that makes a function another app, or
// another app a function: an app becomes a function only when it is created as
// one.
func checkKindCategoryChange(current *entity.AppKindSettings, next base.AppCategory) error {
	wasFunction := current != nil && current.Category == base.AppCategoryFunction
	if wasFunction != (next == base.AppCategoryFunction) {
		return hperrors.Wrap(hperrors.ErrAppKindFunctionUnchangeable)
	}
	return nil
}

// fixFunctionRouting points a function's routing at the port its runtime
// listens on: the app's port, and every domain's.
func fixFunctionRouting(routing *entity.AppRoutingSettings) {
	routing.PinToFunctionPort()
}
