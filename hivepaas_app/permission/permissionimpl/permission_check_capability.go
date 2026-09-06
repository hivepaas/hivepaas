package permissionimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

// capabilityAction is the only action a capability is ever checked for.
//
// A capability names one operation, so "may execute it" is the whole question.
// Setting it here rather than at every call site removes the one thing a caller
// could get wrong about it - and a caller that asked for the wrong action would
// be refused for a reason that has nothing to do with what they hold.
const capabilityAction = base.ActionTypeExecute

func (p *manager) checkCapability(
	ctx context.Context,
	db database.IDB,
	check *permission.CapabilityCheck,
) (bool, error) {
	normalizeCapabilityCheck(check)

	return p.checkFlatResourceAccess(ctx, db, &check.BaseAccessCheck,
		base.ResourceTypeCapability, string(check.Capability))
}

// normalizeCapabilityCheck forces the action the check is answered on.
//
// It overwrites rather than defaults: leaving a caller's value in place would let
// the check answer a different question than its type promises. Clearing AllOf
// and AnyOf as well keeps exactly one of the three set, which is what BaseAccessCheck.IsValid
// requires.
func normalizeCapabilityCheck(check *permission.CapabilityCheck) {
	check.Action, check.AllOf, check.AnyOf = capabilityAction, nil, nil
}
