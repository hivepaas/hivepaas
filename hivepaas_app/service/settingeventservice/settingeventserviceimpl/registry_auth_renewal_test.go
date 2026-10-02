package settingeventserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

// fakeRenewals records what a renewal was recorded for.
type fakeRenewals struct {
	registryauthservice.Service
	auths    [][]string
	keyAuths []string
}

func (f *fakeRenewals) RecordRenewal(_ context.Context, _ database.IDB, authIDs []string) (*entity.Task, error) {
	f.auths = append(f.auths, authIDs)
	return &entity.Task{ID: "renewal"}, nil
}

func (f *fakeRenewals) RecordRenewalForKeyAuth(_ context.Context, _ database.IDB, id string) (*entity.Task, error) {
	f.keyAuths = append(f.keyAuths, id)
	return &entity.Task{ID: "renewal"}, nil
}

func keyAuthSetting(keyID string, status base.SettingStatus) *entity.Setting {
	setting := &entity.Setting{ID: "ka1", Type: base.SettingTypeKeyAuth, Status: status}
	setting.Data = `{"keyId":"` + keyID + `","secretKey":"enc:x"}`
	return setting
}

func ecrCredential(keyAuthID, role string, status base.SettingStatus) *entity.Setting {
	setting := &entity.Setting{ID: "ra1", Type: base.SettingTypeRegistryAuth, Status: status}
	setting.MustSetData(&entity.RegistryAuth{Kind: base.RegistryAuthKindAWSECR, Username: "AWS",
		Address: "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
		ECR: &entity.RegistryAuthECR{Region: "eu-west-1", KeyAuth: entity.ObjectID{ID: keyAuthID},
			RoleARN: role}})
	return setting
}

func onUpdate(t *testing.T, renewals *fakeRenewals, old, updated *entity.Setting, status bool) []*entity.Task {
	t.Helper()
	s := &service{registryAuthService: renewals, settingMountService: &fakeMounts{}}
	event := &settingeventservice.UpdateEvent{Setting: updated, OldSetting: old}
	if status {
		assert.NoError(t, s.OnUpdateStatus(context.Background(), nil, event))
	} else {
		assert.NoError(t, s.OnUpdate(context.Background(), nil, event))
	}
	var renewalTasks []*entity.Task
	for _, task := range event.Tasks {
		if task.ID == "renewal" {
			renewalTasks = append(renewalTasks, task)
		}
	}
	return renewalTasks
}

// New keys in a key auth renew the ECR credentials using it, in the edit's
// transaction; a rename does not, nor does turning it off.
func TestAKeyAuthsNewKeysRenewItsCredentials(t *testing.T) {
	active, disabled := base.SettingStatusActive, base.SettingStatusDisabled

	renewals := &fakeRenewals{}
	tasks := onUpdate(t, renewals, keyAuthSetting("AKIAOLD", active), keyAuthSetting("AKIANEW", active), false)
	assert.Len(t, tasks, 1)
	assert.Equal(t, []string{"ka1"}, renewals.keyAuths)

	renewals = &fakeRenewals{}
	renamed := keyAuthSetting("AKIAOLD", active)
	renamed.Name = "renamed"
	assert.Empty(t, onUpdate(t, renewals, keyAuthSetting("AKIAOLD", active), renamed, false))
	assert.Empty(t, renewals.keyAuths)

	renewals = &fakeRenewals{}
	assert.Empty(t, onUpdate(t, renewals, keyAuthSetting("AKIAOLD", active), keyAuthSetting("AKIAOLD", disabled),
		true))
	assert.Empty(t, renewals.keyAuths, "turned off: a run would fail")

	renewals = &fakeRenewals{}
	assert.Len(t, onUpdate(t, renewals, keyAuthSetting("AKIAOLD", disabled), keyAuthSetting("AKIAOLD", active),
		true), 1)
	assert.Equal(t, []string{"ka1"}, renewals.keyAuths, "turned back on")
}

// An ECR credential linked to another key auth, or assuming another role, is
// renewed; one edited otherwise, or a username and password one, is not.
func TestAnECRCredentialSigningInAnotherWayIsRenewed(t *testing.T) {
	active, disabled := base.SettingStatusActive, base.SettingStatusDisabled

	renewals := &fakeRenewals{}
	assert.Len(t, onUpdate(t, renewals, ecrCredential("ka1", "", active), ecrCredential("ka2", "", active), false), 1)
	assert.Equal(t, [][]string{{"ra1"}}, renewals.auths)

	renewals = &fakeRenewals{}
	assert.Len(t, onUpdate(t, renewals, ecrCredential("ka1", "", active),
		ecrCredential("ka1", "arn:aws:iam::123456789012:role/pull", active), false), 1)

	renewals = &fakeRenewals{}
	assert.Empty(t, onUpdate(t, renewals, ecrCredential("ka1", "", active), ecrCredential("ka1", "", active), false))

	renewals = &fakeRenewals{}
	assert.Len(t, onUpdate(t, renewals, ecrCredential("ka1", "", disabled), ecrCredential("ka1", "", active), true), 1)

	basic := &entity.Setting{ID: "ra2", Type: base.SettingTypeRegistryAuth, Status: active}
	basic.MustSetData(&entity.RegistryAuth{Username: "bot", Address: "ghcr.io"})
	renewals = &fakeRenewals{}
	assert.Empty(t, onUpdate(t, renewals, ecrCredential("ka1", "", active), basic, false))
	assert.Empty(t, renewals.auths)
}
