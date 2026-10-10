package sslcertuc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// checkCustomCert reads a certificate its owner brings: it is a PEM certificate,
// the private key given is its own, and it covers the domain it is set for -
// else the proxy cannot load it, and serves the domain with its default one,
// which no browser trusts. Its expiry is read from it.
func checkCustomCert(cert *entity.SSLCert) error {
	if cert.CertType != base.SSLCertTypeCustom {
		return nil
	}
	block, _ := pem.Decode([]byte(cert.Certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return unusableCert("it is not a PEM certificate")
	}
	leaf, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return unusableCert("it cannot be read: " + err.Error())
	}
	key, err := cert.PrivateKey.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, err = tls.X509KeyPair([]byte(cert.Certificate), []byte(key)); err != nil {
		return unusableCert("the private key is not the certificate's")
	}
	if !certCovers(leaf, cert.Domain) {
		return unusableCert("it does not cover " + cert.Domain)
	}
	cert.ExpireAt = leaf.NotAfter.UTC()
	return nil
}

// certCovers says the certificate is one for the domain: a name of it, or, for
// a wildcard domain, that wildcard.
func certCovers(leaf *x509.Certificate, domain string) bool {
	if strings.HasPrefix(domain, "*.") {
		for _, name := range leaf.DNSNames {
			if strings.EqualFold(name, domain) {
				return true
			}
		}
		return false
	}
	return leaf.VerifyHostname(domain) == nil
}

func unusableCert(reason string) error {
	return hperrors.Wrap(hperrors.ErrSSLCertUnusable).WithParam("Reason", reason)
}
