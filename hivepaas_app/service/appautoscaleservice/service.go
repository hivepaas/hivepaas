// Package appautoscaleservice scales apps' replicas from how busy they are: a
// function from its invocation lines, any other app from its requests in the
// proxy's access log and its containers' CPU in the agent's rows. See
// docs/superpowers/specs/2026-10-03-function-autoscale-design.md and
// 2026-10-03-app-autoscale-design.md.
package appautoscaleservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type Service interface {
	// EnsureJob turns the periodic job that scales apps on while an app has
	// autoscale on, and off when none has: no app with autoscale, no runs at
	// all. It says whether the job changed; the workers are told of that by
	// AnnounceJob, once the change is committed.
	EnsureJob(ctx context.Context, db database.IDB) (changed bool, err error)

	// AnnounceJob tells the workers the job changed, to read it again. Called
	// once the transaction EnsureJob ran in has committed: told before, a
	// worker reads the job as it was, and again only when its cache expires.
	AnnounceJob(ctx context.Context)

	// Run is one run of that job: every app with autoscale on, three queries
	// at most, a decision each. A run that scales one asks for its task to be
	// saved.
	Run(ctx context.Context, data *queue.PeriodicExecData) error

	// Events are an app's scalings since a time, the latest first, at most
	// limit of them (0 for no limit): what the job's saved tasks say of it.
	Events(ctx context.Context, db database.IDB, appID string, since time.Time, limit int) ([]*Event, error)

	// Check says what an app other than a function can scale on now, and
	// what stands in its way. The app comes with its routing setting; svc is
	// its service, nil when it has none.
	Check(ctx context.Context, db database.IDB, app *entity.App, svc *swarm.Service) (*Check, error)
}

// Event is one scaling of an app: when, from and to how many replicas, and
// what it was decided from.
type Event struct {
	Time      time.Time
	From      int
	To        int
	InFlight  float64
	Calls     int64
	Throttled int64
	Requests  int64
	CPU       float64
	Reason    string
}

// The reasons a signal cannot be read, besides the logs' own and the access
// log's.
const (
	// ReasonNotExposed: the app is reached by no domain, so no request of it
	// passes through the proxy.
	ReasonNotExposed = "not-exposed"
	// ReasonAgentUnlabelled: the agent does not mark its rows yet.
	ReasonAgentUnlabelled = "agent-unlabelled"
	// ReasonIdentityMissing: the app's containers do not carry its id, which
	// the agent's rows name it by.
	ReasonIdentityMissing = "identity-missing"
	// ReasonNoCPULimit: the app has neither a CPU limit nor a reservation to
	// measure its CPU by.
	ReasonNoCPULimit = "no-cpu-limit"
)

// The reasons autoscale cannot scale an app at all.
const (
	// RefusedNotReplicated: it does not run a set number of instances.
	RefusedNotReplicated = "not-replicated"
	// RefusedHostPorts: it publishes a port in host mode, one replica a node.
	RefusedHostPorts = "host-ports"
)

// Check is what an app other than a function can scale on now.
type Check struct {
	// Requests and CPU are why a signal cannot be read now, "" when it can.
	Requests string
	CPU      string
	// Refused is why autoscale cannot scale the app at all, "" when it can.
	Refused string
	// Pending is how many of its tasks are wanted and not running: the
	// cluster may have no room for them.
	Pending int
	// WritableMounts is whether its replicas would share, or each have their
	// own, a volume or a bind mount it writes to.
	WritableMounts bool
}
