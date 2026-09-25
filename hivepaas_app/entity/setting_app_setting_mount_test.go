package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func TestAppSettingMountReadsWhatWasStored(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeAppSettingMount, Data: `{"source":{"id":"cert_1"},` +
		`"files":[{"part":"certificate","path":"/etc/app/tls/cert.pem"},` +
		`{"part":"privateKey","path":"/etc/app/tls/key.pem","uid":"1000","mode":"0400"}]}`}

	got, err := setting.AsAppSettingMount()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, &AppSettingMount{
		Source: ObjectID{ID: "cert_1"},
		Files: []*AppSettingMountFile{
			{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
			{Part: "privateKey", Path: "/etc/app/tls/key.pem", UID: "1000", Mode: fileutil.FileMode(0o400)},
		},
	}, got)
}

// The source is a reference like any other: written to res_link, remapped by
// import, and in use while linked.
func TestAppSettingMountReferencesItsSource(t *testing.T) {
	assert.Equal(t, []string{"cert_1"},
		(&AppSettingMount{Source: ObjectID{ID: "cert_1"}}).GetRefObjectIDs().RefSettingIDs)
	assert.Empty(t, (&AppSettingMount{}).GetRefObjectIDs().RefSettingIDs)
}
