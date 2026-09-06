package permissionimpl

import (
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

// The action is the check's own business. A caller that states none - which is
// what every call site now does - must still be answered, and one that states the
// wrong action must not be answered on those terms.
func TestCapabilityCheckSetsItsOwnAction(t *testing.T) {
	tests := []struct {
		name  string
		check *permission.CapabilityCheck
	}{
		{
			name:  "no action stated",
			check: &permission.CapabilityCheck{Capability: base.ResourceCapSecretReveal},
		},
		{
			name: "a wrong action stated is overwritten",
			check: &permission.CapabilityCheck{
				BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeDelete},
				Capability:      base.ResourceCapSecretReveal,
			},
		},
		{
			name: "AllOf and AnyOf are cleared too, so IsValid stays true",
			check: &permission.CapabilityCheck{
				BaseAccessCheck: permission.BaseAccessCheck{
					AllOf: []base.ActionType{base.ActionTypeRead},
					AnyOf: []base.ActionType{base.ActionTypeWrite},
				},
				Capability: base.ResourceCapSecretReveal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeCapabilityCheck(tt.check)

			if tt.check.Action != base.ActionTypeExecute {
				t.Errorf("Action = %q, want %q", tt.check.Action, base.ActionTypeExecute)
			}
			if len(tt.check.AllOf) != 0 || len(tt.check.AnyOf) != 0 {
				t.Error("AllOf and AnyOf must be cleared")
			}
			// Exactly one of the three has to be set, or VerifyAuth rejects it.
			if !tt.check.IsValid() {
				t.Error("the normalized check must be valid")
			}
		})
	}
}

// The action the check asks for and the action a grant is written with have to be
// the same one, or every grant is written and then never matches.
func TestCapabilityActionMatchesTheGrantedAction(t *testing.T) {
	granted := base.AccessActions{Exec: true}
	if !granted.Allows(capabilityAction) {
		t.Errorf("a capability grant carries %+v, which does not allow %q",
			granted, capabilityAction)
	}
}
