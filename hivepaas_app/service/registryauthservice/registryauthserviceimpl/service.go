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
	db            database.IDB
	dockerManager docker.Manager
	appRepo       repository.AppRepo
	settingRepo   repository.SettingRepo
	ecr           ecrTokens
	now           func() time.Time
	// inTx runs fn in a transaction of its own: the database's, or the tests'.
	inTx func(ctx context.Context, fn func(tx database.Tx) error) error
	// interval is the renewal's, as its setting says now.
	interval func(ctx context.Context) time.Duration
}

// New builds the registry auth service. fx wires the arguments from the
// provider list in registry/provides.go.
//
//nolint:ireturn // the constructor of a service returns its interface
func New(
	db *database.DB,
	dockerManager docker.Manager,
	appRepo repository.AppRepo,
	settingRepo repository.SettingRepo,
) registryauthservice.Service {
	s := &service{db: db, dockerManager: dockerManager, appRepo: appRepo, settingRepo: settingRepo,
		ecr: awsECR{}, now: timeutil.NowUTC,
		inTx: func(ctx context.Context, fn func(tx database.Tx) error) error {
			return transaction.Execute(ctx, db, fn)
		}}
	s.interval = s.renewalInterval
	return s
}

// renewalInterval is the renewal setting's interval; the default when it cannot
// be read, which is what a new installation has.
func (s *service) renewalInterval(ctx context.Context) time.Duration {
	if s.settingRepo == nil {
		return entity.RegistryAuthRenewalIntervalDefault
	}
	setting, err := s.settingRepo.GetSingle(ctx, s.db, entity.NewObjectScopeGlobal(),
		base.SettingTypeRegistryAuthRenewal, false)
	if err != nil {
		return entity.RegistryAuthRenewalIntervalDefault
	}
	renewal, err := setting.AsRegistryAuthRenewal()
	if err != nil {
		return entity.RegistryAuthRenewalIntervalDefault
	}
	return renewal.Interval()
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
	token, err := s.ecrTokenFor(ctx, setting, auth, s.interval(ctx))
	if err != nil {
		return nil, err
	}
	return &registry.AuthConfig{Username: ecrUsername, Password: token.password, ServerAddress: auth.Address}, nil
}

// ecrTokenFor is a token for an ECR credential that lives the interval and
// TokenMargin more: the one kept, or one got now and kept.
func (s *service) ecrTokenFor(
	ctx context.Context,
	setting *entity.Setting,
	auth *entity.RegistryAuth,
	interval time.Duration,
) (*ecrToken, error) {
	if token, ok := s.freshToken(auth, interval); ok {
		return token, nil
	}
	return s.renewToken(ctx, setting, interval)
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
func (s *service) freshToken(auth *entity.RegistryAuth, interval time.Duration) (*ecrToken, bool) {
	if auth.Token.IsEmpty() || auth.TokenExpiresAt.Sub(s.now()) < interval+registryauthservice.TokenMargin {
		return nil, false
	}
	token, err := auth.Token.GetPlain()
	if err != nil || token == "" {
		return nil, false
	}
	return &ecrToken{password: token, expiresAt: auth.TokenExpiresAt}, true
}

// renewToken gets a token and keeps it in the credential, in a transaction of
// its own that locks the credential's row: of several processes needing one at
// once, one asks AWS and the others find its token. The token's two fields are
// written alone, without the setting's version, so that a person editing the
// credential does not lose their edit to a version conflict.
func (s *service) renewToken(
	ctx context.Context,
	setting *entity.Setting,
	interval time.Duration,
) (token *ecrToken, err error) {
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
		if kept, ok := s.freshToken(auth, interval); ok { // got by another while this one waited
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
		token = got
		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
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
