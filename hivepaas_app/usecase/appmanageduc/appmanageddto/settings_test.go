package appmanageddto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateManagedSettingsValidation(t *testing.T) {
	const id = "01HZZZZZZZZZZZZZZZZZZZZZZZ"
	for name, tc := range map[string]struct {
		req   *UpdateManagedSettingsReq
		valid bool
	}{
		"a retention":       {&UpdateManagedSettingsReq{RetainFull: new(14)}, true},
		"no retention":      {&UpdateManagedSettingsReq{RetainFull: new(0)}, false},
		"a year and a day":  {&UpdateManagedSettingsReq{RetainFull: new(366)}, false},
		"a cleared query":   {&UpdateManagedSettingsReq{RestoreTestQuery: new("")}, true},
		"a query too long":  {&UpdateManagedSettingsReq{RestoreTestQuery: new(strings.Repeat("x", 2001))}, false},
		"nothing to change": {&UpdateManagedSettingsReq{}, true},
	} {
		tc.req.ProjectID, tc.req.ProjectEnvID, tc.req.AppID = id, id+":prod", id
		assert.Equal(t, tc.valid, len(tc.req.Validate()) == 0, name)
	}
}
