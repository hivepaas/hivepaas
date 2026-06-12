package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

type Client struct {
	client  *lego.Client
	acmeCfg *ACMEConfig
	user    *User
}

type ACMEConfig struct {
	Email          string
	KeyType        base.SSLKeyType
	UserPrivKey    crypto.PrivateKey
	HTTP01Provider challenge.Provider
	HTTP01WebRoot  string
	CADirURL       string // E.g. "https://acme.zerossl.com/v2/DV90"
	EABKid         string
	EABHmacKey     string
}

type User struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.PrivateKey
}

func NewClient(cfg ACMEConfig) (client *Client, err error) {
	if cfg.UserPrivKey == nil {
		cfg.UserPrivKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, apperrors.New(err).WithMsgLog("failed to generate private key for user")
		}
	}

	user := User{
		Email:      cfg.Email,
		PrivateKey: cfg.UserPrivKey,
	}
	legoCfg := lego.NewConfig(&user)
	if legoCfg.CADirURL != "" {
		legoCfg.CADirURL = cfg.CADirURL // Custom ACME endpoint
	}

	switch cfg.KeyType {
	case base.SSLKeyTypeECP256:
		legoCfg.Certificate.KeyType = certcrypto.EC256
	case base.SSLKeyTypeECP384:
		legoCfg.Certificate.KeyType = certcrypto.EC384
	case base.SSLKeyTypeRSA2048:
		legoCfg.Certificate.KeyType = certcrypto.RSA2048
	case base.SSLKeyTypeRSA3072:
		legoCfg.Certificate.KeyType = certcrypto.RSA3072
	case base.SSLKeyTypeRSA4096:
		legoCfg.Certificate.KeyType = certcrypto.RSA4096
	case base.SSLKeyTypeRSA8192:
		legoCfg.Certificate.KeyType = certcrypto.RSA8192
	case base.SSLKeyTypeECP521:
		fallthrough
	default:
		return nil, apperrors.NewUnsupported(fmt.Sprintf("Key type '%v'", cfg.KeyType))
	}

	c, err := lego.NewClient(legoCfg)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create lego client")
	}

	if cfg.HTTP01Provider == nil {
		cfg.HTTP01Provider, err = webroot.NewHTTPProvider(cfg.HTTP01WebRoot)
		if err != nil {
			return nil, apperrors.New(err).WithMsgLog("failed to create http provider for webroot")
		}
	}

	err = c.Challenge.SetHTTP01Provider(cfg.HTTP01Provider)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to set http-01 challenge")
	}

	return &Client{
		client:  c,
		acmeCfg: &cfg,
		user:    &user,
	}, nil
}

func (client *Client) registerUser(_ context.Context) (err error) {
	if client.user.Registration != nil {
		return nil
	}

	var reg *registration.Resource

	// If EAB info is used
	if client.acmeCfg != nil && client.acmeCfg.EABKid != "" && client.acmeCfg.EABHmacKey != "" {
		reg, err = client.client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  client.acmeCfg.EABKid,
			HmacEncoded:          client.acmeCfg.EABHmacKey,
		})
	} else {
		reg, err = client.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
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
) (*certificate.Resource, error) {
	// New users will need to register
	err := client.registerUser(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	certificates, err := client.client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
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
) (*certificate.RenewalInfoResponse, error) {
	// New users will need to register
	err := client.registerUser(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	x509Cert, err := certcrypto.ParsePEMCertificate(cert)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to parse certificate as x509")
	}

	renewalInfo, err := client.client.Certificate.GetRenewalInfo(certificate.RenewalInfoRequest{
		Cert: x509Cert,
	})
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to query renewal info")
	}

	return renewalInfo, nil
}

func (client *Client) ObtainCertificateWithDetails(
	ctx context.Context,
	domains []string,
) (*certificate.Resource, *certificate.RenewalInfoResponse, error) {
	certificates, err := client.ObtainCertificate(ctx, domains)
	if err != nil {
		return nil, nil, apperrors.New(err).WithMsgLog("failed to obtain certificate")
	}

	renewalInfo, err := client.GetRenewalInfo(ctx, certificates.Certificate)
	if err != nil {
		return nil, nil, apperrors.New(err).WithMsgLog("failed to query renewal info")
	}

	return certificates, renewalInfo, nil
}
