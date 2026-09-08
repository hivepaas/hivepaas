package settingsrevertserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Every kind of change that can be put on trial needs a reverter, and the map is
// the only place that is decided. A setting type armed with nothing registered
// here fails its revert at the deadline - which is the worst possible moment,
// because by then the operator is locked out and this was the way back.
//
// Nils are fine: nothing is called, only the wiring is read.
func TestEveryProbationableSettingTypeHasAReverter(t *testing.T) {
	svc := New(nil, nil, nil, nil, nil, nil).(*service)

	for _, typ := range []base.SettingType{
		base.SettingTypeAppRouting,
		base.SettingTypeHivePaaSService,
		base.SettingTypeTraefikConfig,
	} {
		assert.Contains(t, svc.reverters, typ,
			"%s can be put on trial with no way to undo it", typ)
	}
	assert.Len(t, svc.reverters, 3,
		"a reverter was added or removed - check that the setting type arming it is covered")
}
