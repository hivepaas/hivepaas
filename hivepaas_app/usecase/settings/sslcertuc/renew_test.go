package sslcertuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sslservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// renewingSSL obtains a certificate as the real one does - into the setting's
// data - and keeps the settings it wrote Traefik's files of.
type renewingSSL struct {
	sslservice.Service
	written []*entity.Setting
}

func (s *renewingSSL) ObtainCert(
	_ context.Context, setting *entity.Setting, _ *entity.RefObjects, _ bool,
) (bool, error) {
	cert := setting.MustAsSSLCert()
	cert.Certificate = "renewed"
	setting.MustSetData(cert)
	return true, nil
}

func (s *renewingSSL) WriteCertFiles(_ bool, settings ...*entity.Setting) error {
	s.written = append(s.written, settings...)
	return nil
}

type noRefs struct {
	settingservice.Service
}

func (noRefs) LoadRefObjects(context.Context, database.IDB, **entity.RefObjects, *entity.ObjectScope, bool,
	...*entity.Setting) error {
	return nil
}

// A certificate renewed is the one saved, and Traefik's files are written from
// it; the one loaded is left as it was, the record of what the update changed.
func TestARenewedCertificateIsTheOneSaved(t *testing.T) {
	ssl := &renewingSSL{}
	uc := &UC{sslService: ssl, BaseUC: &settings.BaseUC{SettingService: noRefs{}}}
	loaded := &entity.Setting{ID: "c1", Type: base.SettingTypeSSLCert}
	loaded.MustSetData(&entity.SSLCert{CertType: base.SSLCertTypeSelfSigned, Certificate: "old"})
	assert.Equal(t, "old", loaded.MustAsSSLCert().Certificate)
	// UpdateSetting saves a copy of the setting it loaded.
	saving := *loaded

	err := uc.renewInto(context.Background(), database.Tx{}, &entity.ObjectScope{}, &saving)

	assert.NoError(t, err)
	assert.Equal(t, "renewed", saving.MustAsSSLCert().Certificate)
	assert.Equal(t, []*entity.Setting{&saving}, ssl.written)
	assert.Equal(t, "old", loaded.MustAsSSLCert().Certificate)
}
