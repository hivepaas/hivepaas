// Package getstartedservice is what a new installation still has to do, as the
// dashboard's Get started card shows it: the dashboard's certificate, two-factor
// authentication, a GitHub App.
package getstartedservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type ItemStatus string

const (
	ItemStatusTodo      ItemStatus = "todo"
	ItemStatusObtaining ItemStatus = "obtaining"
	ItemStatusFailed    ItemStatus = "failed"
	ItemStatusDone      ItemStatus = "done"
)

// Item is where one thing to do stands. Domain and Error are the dashboard
// certificate's: the name it is for, and why the last attempt failed.
type Item struct {
	Status ItemStatus
	Domain string
	Error  string
}

type Checklist struct {
	DashboardCert Item
	TwoFactor     Item
	GithubApp     Item
}

// AllDone reports whether nothing is left to do.
func (c *Checklist) AllDone() bool {
	return c.DashboardCert.Status == ItemStatusDone && c.TwoFactor.Status == ItemStatusDone &&
		c.GithubApp.Status == ItemStatusDone
}

// CertRequest is what asking for the dashboard's certificate did: the tasks to
// schedule, or, when it asked for nothing, why not - a certificate already
// attached or covering the domain, automatic certificates turned off, a name no
// authority issues for. A button that does nothing without saying so is what
// NotAsked is there to prevent.
type CertRequest struct {
	Tasks    []*entity.Task
	NotAsked string
}

type Service interface {
	// Checklist is each item's state, worked out from what exists. hasTwoFactor
	// is the asking admin's: two-factor authentication is theirs, not the
	// installation's.
	Checklist(ctx context.Context, db database.IDB, hasTwoFactor bool) (*Checklist, error)

	// DashboardCert is the state of the dashboard's certificate alone.
	DashboardCert(ctx context.Context, db database.IDB) (*Item, error)

	// RequestDashboardCert asks for a certificate for the dashboard's domains that
	// have none. The tasks it returns have to be scheduled once the transaction
	// they were written in has committed. ignoreRetryAfter skips the wait a
	// failed attempt leaves: a person asking is reason enough to try again.
	RequestDashboardCert(ctx context.Context, db database.IDB, ignoreRetryAfter bool) (*CertRequest, error)

	// Finish clears the installation step, which hides the card for every admin.
	Finish(ctx context.Context, db database.IDB) error
}
