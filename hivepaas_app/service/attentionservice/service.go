package attentionservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Service finds what, across the cluster, needs someone to look at it: an app
// that is not running, one that keeps restarting, a node that is down.
//
// It answers for the whole cluster and knows nothing of users. Who may see an
// item is decided by its Scope, against the screen the item leads to - the
// caller's to ask.
type Service interface {
	Items(ctx context.Context, db database.IDB) ([]*Item, error)
}

type Kind string

const (
	// KindAppNotRunning is an app with fewer tasks running than it asks for.
	KindAppNotRunning Kind = "app-not-running"
	// KindAppRestarting is an app whose tasks keep failing and being replaced.
	KindAppRestarting Kind = "app-restarting"
	// KindNodeDown is a node the swarm cannot reach.
	KindNodeDown Kind = "node-down"
	// KindNodeOvercommitted is a node whose tasks may use more memory, by their
	// limits, than the node has: when they do, the kernel kills one of them.
	KindNodeOvercommitted Kind = "node-overcommitted"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
)

type ScopeType string

const (
	// ScopeApp is an app of a project, seen by whoever may read its env.
	ScopeApp ScopeType = "app"
	// ScopeSystem is HivePaaS's own apps and services, seen from the system
	// screens.
	ScopeSystem ScopeType = "system"
	// ScopeCluster is the nodes, seen from the cluster screens.
	ScopeCluster ScopeType = "cluster"
)

// Scope is where an item is looked at from, which is who may see it.
type Scope struct {
	Type       ScopeType
	ProjectID  string
	ProjectEnv string // the env's key
	AppID      string
}

type Item struct {
	Kind     Kind
	Severity Severity
	Scope    Scope

	// Subject is what the item is about, named as its screen names it: an
	// app's name, a system service's, a node's hostname.
	Subject     string
	ProjectName string

	// What the item is about, for the screen to word. Each kind fills its own.
	Running      uint64 // tasks running, of an app not running
	Desired      uint64 // tasks asked for
	Restarts     int    // failed tasks in the last hour, of an app restarting
	LastError    string // the latest failed task's error
	NodeState    string // of a node down
	MemoryLimits int64  // the sum of the limits of a node's tasks, in bytes
	MemoryTotal  int64  // the node's memory, in bytes

	// Since is when it started, as near as can be told.
	Since time.Time
}
