package sslservice

import (
	"context"
	"crypto/x509/pkix"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type Service interface {
	WriteCertFiles(forceRecreate bool, settings ...*entity.Setting) error
	DeleteCertFiles(settings ...*entity.Setting) error

	GenerateCertAsPEM(subject *pkix.Name, keyType base.SSLKeyType, notBefore, notAfter time.Time,
		isCA bool) (cert, key []byte, err error)
	ObtainCert(ctx context.Context, sslSetting *entity.Setting, refObjects *entity.RefObjects,
		writeFiles bool) (updated bool, err error)
}
