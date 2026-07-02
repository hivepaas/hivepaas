package sslservice

import (
	"context"
	"crypto/x509/pkix"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/ssl/acme"
)

type Service interface {
	WriteCertFiles(forceRecreate bool, settings ...*entity.Setting) error
	DeleteCertFiles(settings ...*entity.Setting) error

	GenerateCertAsPEM(subject *pkix.Name, keyType base.SSLKeyType, notBefore, notAfter time.Time,
		isCA bool) (cert, key []byte, err error)
	ObtainCert(ctx context.Context, sslSetting *entity.Setting, refObjects *entity.RefObjects,
		writeFiles bool) (updated bool, err error)

	GetAcmeClient(sslSetting *entity.Setting, refObjects *entity.RefObjects) (*acme.Client, error)
}
