package datakeyservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Load installs the data encryption key the app encrypts with, creating one on
	// first start.
	Load(ctx context.Context, db database.IDB) error

	// Rewrap re-encrypts the stored key with a new app secret. This is the whole
	// of an app secret change: the values themselves are untouched.
	Rewrap(ctx context.Context, db database.IDB, currentAppSecret, newAppSecret string) error
}
