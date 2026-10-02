package registryauthserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// ecrUsername is the user every ECR token signs in as.
const ecrUsername = "AWS"

type service struct {
	settingRepo repository.SettingRepo
	ecr         ecrTokens
	now         func() time.Time
	// inTx runs fn in a transaction of its own: the database's, or the tests'.
	inTx func(ctx context.Context, fn func(tx database.Tx) error) error
}

// New builds the registry auth service. fx wires the arguments from the
// provider list in registry/provides.go.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(db *database.DB, settingRepo repository.SettingRepo) registryauthservice.Service {
	return &service{settingRepo: settingRepo, ecr: awsECR{}, now: timeutil.NowUTC,
		inTx: func(ctx context.Context, fn func(tx database.Tx) error) error {
			return transaction.Execute(ctx, db, fn)
		}}
}

func (s *service) AuthHeader(ctx context.Context, setting *entity.Setting) (string, error) {
	auth, err := s.AuthConfig(ctx, setting)
	if err != nil {
		return "", err
	}
	header, err := docker.GenerateAuthHeader(auth)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return header, nil
}

func (s *service) AuthConfig(ctx context.Context, setting *entity.Setting) (*registry.AuthConfig, error) {
	auth, err := setting.AsRegistryAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if auth.Kind != base.RegistryAuthKindAWSECR {
		return basicConfig(auth)
	}
	if token, ok := s.freshToken(auth); ok {
		return &registry.AuthConfig{Username: ecrUsername, Password: token, ServerAddress: auth.Address}, nil
	}

	token, err := s.renewToken(ctx, setting)
	if err != nil {
		return nil, err
	}
	return &registry.AuthConfig{Username: ecrUsername, Password: token, ServerAddress: auth.Address}, nil
}

func (s *service) TryAuth(ctx context.Context, auth *entity.RegistryAuth) (*registry.AuthConfig, error) {
	if auth.Kind != base.RegistryAuthKindAWSECR {
		return basicConfig(auth)
	}
	token, err := s.getToken(ctx, auth)
	if err != nil {
		return nil, err
	}
	return &registry.AuthConfig{Username: ecrUsername, Password: token.password, ServerAddress: auth.Address}, nil
}

func basicConfig(auth *entity.RegistryAuth) (*registry.AuthConfig, error) {
	password, err := auth.Password.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &registry.AuthConfig{Username: auth.Username, Password: password, ServerAddress: auth.Address}, nil
}

// freshToken is the token kept in the credential while it has long enough to
// live to be handed to Swarm.
func (s *service) freshToken(auth *entity.RegistryAuth) (string, bool) {
	if auth.Token.IsEmpty() ||
		auth.TokenExpiresAt.Sub(s.now()) < registryauthservice.RenewalInterval+registryauthservice.TokenMargin {
		return "", false
	}
	token, err := auth.Token.GetPlain()
	if err != nil || token == "" {
		return "", false
	}
	return token, true
}

// renewToken gets a token and keeps it in the credential, in a transaction of
// its own that locks the credential's row: of several processes needing one at
// once, one asks AWS and the others find its token. The token's two fields are
// written alone, without the setting's version, so that a person editing the
// credential does not lose their edit to a version conflict.
func (s *service) renewToken(ctx context.Context, setting *entity.Setting) (token string, err error) {
	err = s.inTx(ctx, func(db database.Tx) error {
		current, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeRegistryAuth, setting.ID, false,
			bunex.SelectFor("UPDATE"))
		if err != nil {
			return hperrors.Wrap(err)
		}
		auth, err := current.AsRegistryAuth()
		if err != nil {
			return hperrors.Wrap(err)
		}
		if kept, ok := s.freshToken(auth); ok { // got by another while this one waited
			token = kept
			return nil
		}
		got, err := s.getToken(ctx, auth)
		if err != nil {
			return err
		}
		auth.Token = entity.NewEncryptedField(got.password)
		auth.TokenExpiresAt = got.expiresAt
		if err = current.SetData(auth); err != nil {
			return hperrors.Wrap(err)
		}
		if err = s.settingRepo.Update(ctx, db, current, bunex.UpdateColumns("data")); err != nil {
			return hperrors.Wrap(err)
		}
		token = got.password
		return nil
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return token, nil
}

// getToken asks AWS for a token with the credential's keys.
func (s *service) getToken(ctx context.Context, auth *entity.RegistryAuth) (*ecrToken, error) {
	if auth.ECR == nil {
		return nil, hperrors.NewArgumentInvalid("registry credential").
			WithExtraDetail("The Amazon ECR credential for %s has no AWS keys.", auth.Address)
	}
	secret, err := auth.ECR.SecretAccessKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	token, err := s.ecr.Token(ctx, auth.ECR, secret)
	if err != nil {
		return nil, hperrors.NewArgumentInvalid("registry credential").WithCause(err).
			WithExtraDetail("The AWS keys of the Amazon ECR credential for %s were refused: %v", auth.Address, err)
	}
	if token.expiresAt.IsZero() {
		token.expiresAt = s.now().Add(ecrTokenLife)
	}
	return token, nil
}

// ecrTokenLife is how long ECR says a token lives, for an answer that does not.
const ecrTokenLife = 12 * time.Hour
