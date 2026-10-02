package entity

import (
	"errors"
	"regexp"
	"time"

	"github.com/moby/moby/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	CurrentRegistryAuthVersion = 1
)

var errECRNeedsToken = errors.New("an Amazon ECR credential signs in with a token got for it, " +
	"not a password: use the registry auth service")

var _ = registerSettingParser(base.SettingTypeRegistryAuth, &registryAuthParser{})

type registryAuthParser struct {
}

func (s *registryAuthParser) New() SettingData {
	return &RegistryAuth{}
}

// RegistryAuthManagedBySystemRegistry marks, in RegistryAuth.ManagedBy, the
// credential the system registry created for itself. It is how the registry
// knows its own credential again whatever it is called or addressed at now: an
// operator may rename it, and the registry's domain may change between one
// switch-on and the next. It is in the data rather than in Setting.Kind, which
// a registry auth keeps its address in.
const RegistryAuthManagedBySystemRegistry = "system-registry"

type RegistryAuth struct {
	// Kind is how it signs in: a username and a password (empty), or AWS keys
	// for Amazon ECR, kept in a key auth.
	Kind     base.RegistryAuthKind `json:"kind,omitempty"`
	Username string                `json:"username"`
	Password EncryptedField        `json:"password"`
	Address  string                `json:"address"`
	Readonly bool                  `json:"readonly,omitempty"`
	// ManagedBy names what created the credential and keeps it, when that is not
	// an operator. It cannot be set or cleared through the API: an edit carries
	// it over.
	ManagedBy string `json:"managedBy,omitempty"`
	// ECR is the AWS side of an Amazon ECR credential.
	ECR *RegistryAuthECR `json:"ecr,omitempty"`
	// Token is the last ECR token got for the credential, and TokenExpiresAt
	// when it stops working. They are derived from the keys: never answered by
	// the API, never exported, and cleared when the credential's ECR side
	// changes.
	Token          EncryptedField `json:"token,omitzero"`
	TokenExpiresAt time.Time      `json:"tokenExpiresAt,omitzero"`
	// TokenKeyVer is the key auth's version the token was got with: a token got
	// with keys since edited is not used.
	TokenKeyVer int `json:"tokenKeyVer,omitzero"`
}

// RegistryAuthECR is what signs in to Amazon ECR: the AWS keys of a key auth,
// and a role they assume first when one is named. The registry's account and
// region are in its address, <account>.dkr.ecr.<region>.amazonaws.com.
type RegistryAuthECR struct {
	Region  string   `json:"region"`
	KeyAuth ObjectID `json:"keyAuth"`
	RoleARN string   `json:"roleArn,omitempty"`
}

// ecrAddressRegexes are an ECR registry's addresses, each giving its account
// and its region: <account>.dkr.ecr[-fips].<region>.amazonaws.com[.cn], and
// the dual-stack <account>.dkr-ecr.<region>.on.aws.
var ecrAddressRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^([0-9]{12})\.dkr\.ecr(?:-fips)?\.([a-z]{2}(?:-[a-z]+)+-[0-9])\.amazonaws\.com(?:\.cn)?$`),
	regexp.MustCompile(`^([0-9]{12})\.dkr-ecr\.([a-z]{2}(?:-[a-z]+)+-[0-9])\.on\.aws$`),
}

// ParseECRAddress is the account and the region of an ECR registry's address;
// ok is false for an address that is not one. It is what keeps a token on its
// way to AWS: Docker hands the token to the address, so an address that were
// anyone's would be given the token of a key auth it may not read.
func ParseECRAddress(address string) (account, region string, ok bool) {
	for _, re := range ecrAddressRegexes {
		if m := re.FindStringSubmatch(address); m != nil {
			return m[1], m[2], true
		}
	}
	return "", "", false
}

// SameECRKeys reports whether two credentials sign in to ECR the same way: a
// token got with one is good for the other. The key auth's own edits are told
// apart by TokenKeyVer.
func (s *RegistryAuth) SameECRKeys(other *RegistryAuth) bool {
	if s.ECR == nil || other == nil || other.ECR == nil || s.Address != other.Address {
		return false
	}
	return s.ECR.KeyAuth.ID == other.ECR.KeyAuth.ID && s.ECR.RoleARN == other.ECR.RoleARN &&
		s.ECR.Region == other.ECR.Region
}

func (s *RegistryAuth) GetType() base.SettingType {
	return base.SettingTypeRegistryAuth
}

// GetRefObjectIDs is the key auth of an ECR credential: linked, it is shown as
// in use, and saving checks the credential's scope can see it.
func (s *RegistryAuth) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.ECR != nil && s.ECR.KeyAuth.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.ECR.KeyAuth.ID)
	}
	return refIDs
}

func (s *RegistryAuth) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt reveals the credential's password. An ECR credential's keys are the
// key auth's to reveal, and the token is never answered.
func (s *RegistryAuth) Decrypt() error {
	_, err := s.Password.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// GenerateAuthHeader is the credential as Docker's X-Registry-Auth, for a
// username and a password. An ECR credential has no password to give: its token
// is got by the registry auth service, which every use goes through.
func (s *RegistryAuth) GenerateAuthHeader() (string, error) {
	if s.Kind == base.RegistryAuthKindAWSECR {
		return "", hperrors.NewInfra(errECRNeedsToken)
	}
	password, err := s.Password.GetPlain()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	h, err := docker.GenerateAuthHeader(&registry.AuthConfig{
		Username:      s.Username,
		Password:      password,
		ServerAddress: s.Address,
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return h, nil
}

func (s *Setting) AsRegistryAuth() (*RegistryAuth, error) {
	return parseSettingAs[*RegistryAuth](s)
}

func (s *Setting) MustAsRegistryAuth() *RegistryAuth {
	return gofn.Must(s.AsRegistryAuth())
}
