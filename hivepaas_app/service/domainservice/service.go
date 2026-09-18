package domainservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	VerifyProjectDomains(ctx context.Context, db database.IDB, projectID string, domains []string) error
	VerifyDomainsAvailable(ctx context.Context, db database.IDB, domains []string, ignoreAppIDs []string) error

	// FindCertsForDomains maps each domain to the certificate it can be served
	// with: one issued for exactly that name, or a wildcard covering it. A domain
	// no certificate in the scope covers is absent from the result, which is not
	// an error - an app can be routed without one.
	FindCertsForDomains(ctx context.Context, db database.IDB, scope *entity.ObjectScope,
		domains []string) (map[string]*entity.Setting, error)
}
