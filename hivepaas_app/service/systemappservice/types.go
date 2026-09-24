package systemappservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// DefaultEnv is the environment a system app lands in when it names none - the
// one the stack's own services are synced into.
const DefaultEnv = "default"

type ProvisionReq struct {
	// Env is the hivepaas project environment the app goes in, created when it
	// does not exist. An environment is a network, so a system app that must not
	// share one with the others gets an environment of its own.
	Env  string
	Key  string
	Name string
	Note string

	// Doc is built the way a template's app is, so a system app cannot drift away
	// from what an app can be.
	Doc *specmodel.AppDoc

	// Customize writes onto the service what a document is not allowed to say: a
	// mode, a bind mount of a host path, another network, a log driver. It runs
	// after the document is built, so it sees - and may change - what the build
	// wrote. Those limits exist for templates, which are written by somebody
	// else; a system app is written here.
	Customize func(spec *swarm.ServiceSpec) error

	TriggerUserID string
}

type ProvisionResp struct {
	App *entity.App

	// Unscheduled: see Service.Provision.
	DeploymentTask *entity.Task
	CertTasks      []*entity.Task

	// Cleanup removes from docker what provisioning created there. The caller
	// runs it when its transaction does not commit: the app's records go with the
	// transaction, and nothing else would take the service down. It is set even
	// when Provision fails, so that a half-made app is not left running.
	Cleanup func(ctx context.Context) error
}

type RedeployReq struct {
	// App has its Settings loaded, as LoadApp returns it.
	App *entity.App

	// Change edits the deployment settings and says whether it changed anything.
	// Nothing is written or deployed when it did not.
	Change func(settings *entity.AppDeploymentSettings) bool

	TriggerUserID string
}

// SecretFile is one secret an app reads from a file, so that its value never
// appears in the service spec the way an argument or an environment variable
// does.
type SecretFile struct {
	// Key names the secret among the app's secrets.
	Key string
	// Path is where the file is mounted in the container.
	Path  string
	Value string
}
