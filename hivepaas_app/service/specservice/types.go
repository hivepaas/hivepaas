// Package specservice exports HivePaaS configuration as a portable bundle.
//
// A spec is a faithful snapshot: it records values that are meaningful only on
// the installation that produced it, because the common case is re-import into
// that same installation. Portability is handled at import time by a
// skip-missing flag rather than by weakening what export records.
package specservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type ExportReq struct {
	Scope       *entity.ObjectScope
	SecretsMode specmodel.SecretsMode
	// Passphrase is required when SecretsMode is encrypted.
	Passphrase string
	// WorkDir is where the bundle is staged and archived. The caller owns it and
	// is responsible for removing it once the response has been read.
	WorkDir string
}

type ExportResp struct {
	// Path is the finished archive on disk, inside WorkDir.
	Path string
	// Filename is what the download should be called.
	Filename string
	Size     int64
	// Report is the full detail, also written into the bundle as report.yaml.
	Report *specmodel.Report
	// Summary is what a response header can carry.
	Summary *specmodel.ReportSummary
}

// BuildAppReq asks for an AppDoc to be built into an app being provisioned.
type BuildAppReq struct {
	// App is the app being provisioned: ID, Key, Project and ProjectEnv set.
	App *entity.App
	Doc *specmodel.AppDoc
	// Spec is the app's initial service spec. The build writes into it.
	Spec    *swarm.ServiceSpec
	TimeNow time.Time
	// Import says the document is an export rather than a template: it is checked
	// with CheckImportable, every block export writes is built, and each block
	// replaces what Spec holds.
	Import bool
}

type BuildAppResp struct {
	// Settings replace the app's default settings of the same type.
	Settings []*entity.Setting
}

// ValidateImportReq asks what importing a bundle at a scope would do.
type ValidateImportReq struct {
	// Scope is the route's: global, a project, or a project env. It bounds what
	// the import may write.
	Scope *entity.ObjectScope
	// Bundle is the archive as export produced it, age-encrypted or not.
	Bundle     []byte
	Passphrase string
	Selection  specmodel.Selection
	Options    specmodel.ImportOptions
	// AuthorizeSecrets is asked, once the bundle is read, whether this caller may
	// have secrets revealed. It is asked only for a bundle that carries them,
	// since planning it compares this installation's secrets with the bundle's,
	// and before anything is compared. Nil allows it.
	AuthorizeSecrets func(ctx context.Context, mode specmodel.SecretsMode) error
	// MayWriteCluster answers whether this caller has Write on the cluster
	// module, which giving an app more of the host than a container ordinarily
	// gets takes: capabilities, and mounts of the host's own paths and volumes.
	// It is asked at most once, and only when an imported app would. Nil allows
	// it.
	MayWriteCluster func(ctx context.Context) (bool, error)
	// MayWriteApp answers whether this caller may write to an app, which an
	// imported mount reaching that app's storage needs. It is asked only for an
	// app the import does not itself write. Nil allows it.
	MayWriteApp func(ctx context.Context, app *entity.App) (bool, error)
	// MayChangeOwner answers whether this caller may give an existing project
	// another owner. It is asked only for a project whose owner the bundle
	// changes. Nil allows it.
	MayChangeOwner func(ctx context.Context, project *entity.Project) (bool, error)
}

// ApplyImportReq applies the plan a ValidateImportReq made. PlanHash is the
// hash of the plan the operator saw: apply plans the request again, and refuses
// a plan that is not the same.
type ApplyImportReq struct {
	ValidateImportReq
	// OperatorID owns a project created for an owner nobody here is.
	OperatorID   string
	PlanHash     string
	AcceptIssues bool
}

type ApplyImportResp struct {
	// Plan is the plan applied, with each selected node's outcome.
	Plan *specmodel.ImportPlan
	// Cleanup removes from docker what provisioning the new apps created. It is
	// the caller's to run when its transaction did not commit: the records are
	// gone, and the services, secrets and configs are not. Its context should be
	// one that is not already canceled.
	Cleanup func(ctx context.Context) error
	// AfterCommit is phase 2: what writing takes beyond the database - the
	// running services of the apps updated, their swarm files, routing and
	// environment, certificate files. It runs once the caller's transaction has
	// committed, on db, and records each app's outcome in Plan. An app that fails
	// does not stop the others.
	AfterCommit func(ctx context.Context, db database.IDB) error
	// Tasks are the deployments and certificate tasks phase 1 created. They can
	// be scheduled once the transaction has committed, and after AfterCommit, so
	// that a deployment runs on the service phase 2 updated.
	Tasks []*entity.Task
	// Deployments are the deployments queued.
	Deployments []*ImportDeployment
}

// ImportDeployment is a deployment an import queued.
type ImportDeployment struct {
	AppID        string `json:"appId"`
	DeploymentID string `json:"deploymentId"`
}

type ValidateImportResp struct {
	Plan *specmodel.ImportPlan
	// SecretsMode is the bundle's, which decides the permission reading it needs.
	SecretsMode specmodel.SecretsMode
}
