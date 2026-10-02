// Package registryauthservice answers a registry credential as Docker takes it.
// A username and a password are used as stored; an Amazon ECR credential's
// password is a token that expires after 12 hours, got from its AWS keys when
// one is needed and kept in the credential until it is too old to hand over.
package registryauthservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// RenewalInterval is how often the credentials Swarm keeps in its services are
// renewed. A token handed to Swarm must live until the next renewal: the one
// kept is used while it has this, and TokenMargin, left.
const RenewalInterval = 6 * time.Hour

// TokenMargin is what a token must live past the next renewal.
const TokenMargin = time.Hour

type Service interface {
	// AuthConfig is the credential as Docker takes it. For an ECR credential
	// its password is a token that lives at least RenewalInterval and
	// TokenMargin more: the one kept, or one got now and kept.
	AuthConfig(ctx context.Context, setting *entity.Setting) (*registry.AuthConfig, error)
	// AuthHeader is AuthConfig encoded as Docker's X-Registry-Auth; empty for a
	// credential with no password.
	AuthHeader(ctx context.Context, setting *entity.Setting) (string, error)
	// TryAuth is AuthConfig for a credential not saved, such as one being
	// tested: an ECR token is got, and kept nowhere.
	TryAuth(ctx context.Context, auth *entity.RegistryAuth) (*registry.AuthConfig, error)
}
