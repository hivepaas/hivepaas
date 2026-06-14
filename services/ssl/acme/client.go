package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

type Client struct {
	client  *lego.Client
	acmeCfg *ACMEConfig
	user    *User
}

type ACMEConfig struct {
	CACode string
	Email  string

	EABKid     string
	EABHmacKey string

	HTTP01Provider challenge.Provider
	DNS01Provider  challenge.Provider
}

type User struct {
	registration.User
	Email        string
	Registration *acme.ExtendedAccount
	PrivateKey   crypto.Signer
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *acme.ExtendedAccount {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.Signer {
	return u.PrivateKey
}

func NewClient(cfg ACMEConfig) (client *Client, err error) {
	userPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to generate private key for user")
	}

	user := &User{
		Email:      cfg.Email,
		PrivateKey: userPrivKey,
	}
	legoCfg := lego.NewConfig(user)

	if cfg.CACode != "" {
		legoCfg.CADirURL, err = lego.GetDirectoryURL(cfg.CACode)
		if err != nil {
			return nil, apperrors.New(err).
				WithMsgLog("failed to get directory URL for CA code '%s'", cfg.CACode)
		}
	}

	c, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create lego client")
	}

	switch {
	case cfg.HTTP01Provider != nil:
		err = c.Challenge.SetHTTP01Provider(cfg.HTTP01Provider)
		if err != nil {
			return nil, apperrors.New(err).WithMsgLog("failed to set http-01 challenge")
		}
	case cfg.DNS01Provider != nil:
		err = c.Challenge.SetDNS01Provider(cfg.DNS01Provider)
		if err != nil {
			return nil, apperrors.New(err).WithMsgLog("failed to set dns-01 challenge")
		}
	default:
		return nil, apperrors.NewMissing("ACME challenge provider")
	}

	return &Client{
		client:  c,
		acmeCfg: &cfg,
		user:    user,
	}, nil
}

func (client *Client) registerUser(ctx context.Context) (err error) {
	if client.user.Registration != nil {
		return nil
	}

	var reg *acme.ExtendedAccount

	// If EAB info is used
	if client.acmeCfg != nil && client.acmeCfg.EABKid != "" && client.acmeCfg.EABHmacKey != "" {
		reg, err = client.client.Registration.RegisterWithExternalAccountBinding(ctx, registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  client.acmeCfg.EABKid,
			HmacEncoded:          client.acmeCfg.EABHmacKey,
		})
	} else {
		reg, err = client.client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	}
	if err != nil {
		return apperrors.New(err).WithMsgLog("failed to register user")
	}
	client.user.Registration = reg

	return nil
}

func (client *Client) ObtainCertificate(
	ctx context.Context,
	domains []string,
	keyType base.SSLKeyType,
) (*certificate.Resource, error) {
	// New users will need to register
	err := client.registerUser(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	certKeyType, err := client.getKeyType(keyType)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	certificates, err := client.client.Certificate.Obtain(ctx, certificate.ObtainRequest{
		Domains: domains,
		KeyType: certKeyType,
		Bundle:  true,
	})
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to obtain certificate")
	}

	return certificates, nil
}

func (client *Client) GetRenewalInfo(
	ctx context.Context,
	cert []byte,
) (*certificate.RenewalInfo, error) {
	// New users will need to register
	err := client.registerUser(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	x509Cert, err := certcrypto.ParsePEMCertificate(cert)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to parse certificate as x509")
	}

	renewalInfo, err := client.client.Certificate.GetRenewalInfo(ctx, x509Cert)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to query renewal info")
	}

	return renewalInfo, nil
}

func (client *Client) ObtainCertificateWithDetails(
	ctx context.Context,
	domains []string,
	keyType base.SSLKeyType,
) (*certificate.Resource, *certificate.RenewalInfo, error) {
	certificates, err := client.ObtainCertificate(ctx, domains, keyType)
	if err != nil {
		return nil, nil, apperrors.New(err).WithMsgLog("failed to obtain certificate")
	}

	renewalInfo, err := client.GetRenewalInfo(ctx, certificates.Certificate)
	if err != nil {
		return nil, nil, apperrors.New(err).WithMsgLog("failed to query renewal info")
	}

	return certificates, renewalInfo, nil
}

func (client *Client) getKeyType(
	keyType base.SSLKeyType,
) (certcrypto.KeyType, error) {
	switch keyType {
	case base.SSLKeyTypeECP256:
		return certcrypto.EC256, nil
	case base.SSLKeyTypeECP384:
		return certcrypto.EC384, nil
	case base.SSLKeyTypeRSA2048:
		return certcrypto.RSA2048, nil
	case base.SSLKeyTypeRSA3072:
		return certcrypto.RSA3072, nil
	case base.SSLKeyTypeRSA4096:
		return certcrypto.RSA4096, nil
	case base.SSLKeyTypeRSA8192:
		return certcrypto.RSA8192, nil
	case base.SSLKeyTypeECP521:
		fallthrough
	default:
		return "", apperrors.NewUnsupported(fmt.Sprintf("Key type '%v'", keyType))
	}
}
