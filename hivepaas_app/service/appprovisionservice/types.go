package appprovisionservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// ConfigureFunc fills in an app's configuration. It runs once the app and its
// initial service spec are prepared and before the service is created. The
// settings it returns replace the app's default settings of the same type; what
// it writes into spec is created with the service.
type ConfigureFunc func(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) (
	[]*entity.Setting, error)

type ProvisionAppReq struct {
	ProjectID    string
	ProjectEnvID string
	// AppID is the id the app is created with, generated when empty. The apps a
	// template creates together name each other by id, so their ids are chosen
	// before the first of them exists.
	AppID  string
	Name   string
	Status base.AppStatus
	Note   string
	Tags   []string
	// LogicalParentID is the app this one is being created to serve, empty for an
	// app created on its own. A template's dependencies carry the id of the app
	// they were created with.
	LogicalParentID string
	// Configure is nil for an empty app.
	Configure ConfigureFunc
	// Deployment asks for the deployment that replaces the placeholder image with
	// the one the app's deployment settings name. It is nil for an empty app,
	// which has nothing to deploy yet.
	Deployment *FirstDeployment
}

// FirstDeployment is what an app's first deployment records about who asked for
// it. The app is new, so there is nothing else to say about it.
type FirstDeployment struct {
	Source   base.DeploymentTriggerSource
	SourceID string
}

type ProvisionAppResp struct {
	// App has Project, ProjectEnv and Settings set, and the ServiceID of the
	// service created for it.
	App *entity.App
	// Deployment and DeploymentTask are set when the request asked for a first
	// deployment.
	//
	// The task is created but not scheduled: a task row can be picked up only
	// once the transaction it was written in has committed, which is the
	// caller's to wait for.
	Deployment     *entity.Deployment
	DeploymentTask *entity.Task
	// CertTasks obtain the certificates the app's domains have none for, and are
	// unscheduled for the same reason.
	CertTasks []*entity.Task
	// Created is what provisioning made in docker, for a caller undoing it.
	Created *CreatedInDocker
}

// CreatedInDocker is what exists in docker for an app but not in the database,
// so that a transaction that rolls back can be followed by removing it.
type CreatedInDocker struct {
	ServiceID string
	Secrets   []*entity.SwarmSecretRef
	Configs   []*entity.SwarmConfigRef
}

type ProvisionAppsReq struct {
	// Apps are provisioned in the order given.
	Apps []*ProvisionAppReq
}

type ProvisionAppsResp struct {
	// Apps is what was provisioned, in request order. It is set even when
	// provisioning failed part way, and then holds the apps created before it.
	Apps []*ProvisionAppResp
	// Cleanup removes from docker everything the request created, newest first.
	// Call it when the transaction the request ran in did not commit: the records
	// are gone, and what was created in docker is not.
	//
	// Its context is the caller's to choose, and it should be one that is not
	// already canceled - a request that failed on its way out often has one.
	Cleanup func(ctx context.Context) error
}

type ApplyAppConfigurationReq struct {
	// App is an app whose settings are already persisted.
	App *entity.App
	// RefObjects are the objects the app's settings refer to. An empty set is
	// used when it is nil, which is what a new app has: nothing has been loaded
	// for it yet.
	RefObjects *entity.RefObjects
}

type ApplyAppConfigurationResp struct {
	// Secrets and Configs are the docker objects created, in the order the app's
	// settings list them. An entry is nil where the setting asked for no file to
	// be mounted, which is a secret read through the environment.
	Secrets []*entity.SwarmSecretRef
	Configs []*entity.SwarmConfigRef
	// CertTasks obtain the certificates the app's domains have none for. Like
	// DeploymentTask they are created but not scheduled - see ProvisionAppResp.
	CertTasks []*entity.Task
}
