package apikeydto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func newCreateReq(access *base.AccessActions) *CreateAPIKeyReq {
	req := NewCreateAPIKeyReq()
	req.Name = "ci"
	req.ExpireAt = timeutil.NowUTC().Add(24 * time.Hour)
	req.AccessAction = access
	return req
}

func hasFieldError(errs hperrors.ValidationErrors, field string) bool {
	for _, err := range errs {
		if err != nil && strings.Contains(err.Error(), field) {
			return true
		}
	}
	return false
}

func TestCreateAPIKeyRequiresAccessAction(t *testing.T) {
	tests := []struct {
		name    string
		access  *base.AccessActions
		wantErr bool
	}{
		{
			// Omitting it used to mean "no limit", so the key silently carried the
			// owner's full authority - the one outcome nobody picks on purpose.
			name:    "omitted is refused",
			access:  nil,
			wantErr: true,
		},
		{
			// A key that may do nothing cannot do any work: a mistake, like the above.
			name:    "granting nothing is refused",
			access:  &base.AccessActions{},
			wantErr: true,
		},
		{
			name:   "a narrowed key is accepted",
			access: &base.AccessActions{Read: true},
		},
		{
			// Full access stays allowed - stated explicitly, which is the point.
			name:   "full access is accepted when it is asked for",
			access: &base.AccessActions{Read: true, Exec: true, Write: true, Del: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := newCreateReq(tt.access).Validate()
			if got := hasFieldError(errs, "accessAction"); got != tt.wantErr {
				t.Errorf("accessAction rejected = %v, want %v (errs: %v)", got, tt.wantErr, errs)
			}
		})
	}
}

// A key with no stored limit carries its owner's full authority. Reporting it as
// a value made it read as granting nothing at all - the most dangerous key on the
// list looking like the safest.
func TestUnlimitedKeyIsNotReportedAsHavingNoAccess(t *testing.T) {
	setting := &entity.Setting{ID: "set_1", Name: "legacy", Type: base.SettingTypeAPIKey}
	if err := setting.SetData(&entity.APIKey{KeyID: "k1", AccessAction: nil}); err != nil {
		t.Fatal(err)
	}

	resp, err := TransformAPIKey(setting, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessAction != nil {
		t.Errorf("an unlimited key must not report a set of actions, got %+v", *resp.AccessAction)
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), `"accessAction":{"read":false`) {
		t.Errorf("an unlimited key is reported as granting nothing: %s", body)
	}
}

func TestNarrowedKeyKeepsItsActions(t *testing.T) {
	setting := &entity.Setting{ID: "set_2", Name: "ro", Type: base.SettingTypeAPIKey}
	if err := setting.SetData(&entity.APIKey{KeyID: "k2",
		AccessAction: &base.AccessActions{Read: true}}); err != nil {
		t.Fatal(err)
	}

	resp, err := TransformAPIKey(setting, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessAction == nil || !resp.AccessAction.Read || resp.AccessAction.Write {
		t.Errorf("wrong actions reported: %+v", resp.AccessAction)
	}
}
