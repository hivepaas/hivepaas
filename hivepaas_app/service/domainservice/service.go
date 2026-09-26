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

	// EnsureCertsForDomains gives every domain a certificate: the one the system
	// already holds, or one it starts obtaining in the background.
	//
	// What comes back in Matched can be attached now. What comes back in
	// Obtaining has a task behind it and nothing to serve yet; the task attaches
	// it and applies the routing again when there is. Skipped says, per domain,
	// why neither happened - a local name no authority issues for, automatic
	// certificates turned off, an attempt already in flight.
	EnsureCertsForDomains(ctx context.Context, db database.IDB,
		req *EnsureCertsReq) (*EnsureCertsResp, error)
}

type EnsureCertsReq struct {
	// Scope is where certificates are looked for: an app's scope sees its
	// project's and the installation's as well.
	Scope     *entity.ObjectScope
	ProjectID string
	// AppID is whose routing gets applied again once a certificate arrives.
	AppID   string
	Domains []string
	// IgnoreRetryAfter asks again for a certificate a failed attempt left waiting:
	// a person asking is reason enough. One being obtained is still not asked for
	// twice by the caller, which checks first.
	IgnoreRetryAfter bool
	// IgnoreSelfSigned leaves a self-signed certificate that covers a domain out
	// of the matching, so one a browser trusts is asked for instead. The one an
	// installation signs itself is made for its root domain, which the dashboard's
	// domain can be.
	IgnoreSelfSigned bool
}

type EnsureCertsResp struct {
	Matched   map[string]*entity.Setting
	Obtaining []*entity.Setting
	Skipped   map[string]string
	// Tasks are the obtain tasks written for Obtaining. They are created but not
	// scheduled: a task row can be picked up only once the transaction it was
	// written in has committed, which is the caller's to wait for. A caller that
	// schedules none of them loses nothing but time - the queue's own scan finds
	// them at its next pass.
	Tasks []*entity.Task
}
