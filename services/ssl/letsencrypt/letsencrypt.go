package letsencrypt

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

type Client struct {
	client *lego.Client
	user   *User
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

func NewClient(email string, keyType base.SSLKeyType, http01NginxRoot string) (client *Client, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to generate private key for user")
	}

	user := User{
		Email:      email,
		PrivateKey: privateKey,
	}
	cfg := lego.NewConfig(&user)

	switch keyType { //nolint:exhaustive
	case base.SSLKeyTypeECP256:
		cfg.Certificate.KeyType = certcrypto.EC256
	case base.SSLKeyTypeECP384:
		cfg.Certificate.KeyType = certcrypto.EC384
	case base.SSLKeyTypeRSA2048:
		cfg.Certificate.KeyType = certcrypto.RSA2048
	case base.SSLKeyTypeRSA3072:
		cfg.Certificate.KeyType = certcrypto.RSA3072
	case base.SSLKeyTypeRSA4096:
		cfg.Certificate.KeyType = certcrypto.RSA4096
	default:
		return nil, apperrors.NewUnsupported(fmt.Sprintf("Key type '%v'", keyType))
	}

	c, err := lego.NewClient(cfg)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create lego client")
	}

	webrootProvider, err := webroot.NewHTTPProvider(http01NginxRoot)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to create http provider for webroot")
	}

	err = c.Challenge.SetHTTP01Provider(webrootProvider)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to set http-01 challenge")
	}

	return &Client{
		client: c,
		user:   &user,
	}, nil
}

func (client *Client) registerUser(_ context.Context) error {
	if client.user.Registration != nil {
		return nil
	}
	reg, err := client.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
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
