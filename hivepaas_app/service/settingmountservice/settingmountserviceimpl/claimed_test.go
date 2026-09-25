package settingmountserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Only entries give an app files: every path they claim, but the entry being
// saved and a disabled one.
func TestClaimedPathsAreTheEntriesPaths(t *testing.T) {
	useDataKey(t)
	saved := entry(t, "saved", base.SettingStatusActive, certFiles("cert_1"))
	disabled := entry(t, "off", base.SettingStatusDisabled, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "certificate", Path: "/etc/off.pem"}}})
	other := entry(t, "other", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "certificate", Path: "/etc/other.pem"}}})
	svc := &service{}
	svc.loadClaimants = func(context.Context, database.IDB, string) ([]*entity.Setting, error) {
		return []*entity.Setting{saved, disabled, other}, nil
	}

	claimed, err := svc.ClaimedPaths(context.Background(), nil, testApp.ID, saved.ID)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"/etc/other.pem": "setting mount other"}, claimed)
}
