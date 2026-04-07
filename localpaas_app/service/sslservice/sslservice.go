package sslservice

import (
	"crypto/x509/pkix"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

type SSLService interface {
	GenerateCert(subject *pkix.Name, keyType base.SSLKeyType, notBefore, notAfter time.Time,
		isCA bool) (cert, key []byte, err error)
}

func NewSSLService() SSLService {
	return &sslService{}
}

type sslService struct {
}
