package entity

import (
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// hivepaasStackAppKeys are the apps of the stack HivePaaS is deployed as: what
// runs this installation. The apps HivePaaS provisions besides - the logs'
// backend and collector, the registry - are turned off with their feature.
var hivepaasStackAppKeys = []string{
	base.HivepaasAppKey, base.HivepaasWorkerKey, base.HivepaasDbKey, base.HivepaasCacheKey,
	base.HivepaasTraefikKey, base.HivepaasUpdaterKey, base.HivepaasDockerProxyKey, base.HivepaasAgentKey,
}

func isHivepaasProject(project *Project) bool {
	return project != nil && project.Key == base.HivepaasProjectKey
}

// CheckProjectChange refuses deleting or disabling the project HivePaaS runs
// in - action says which, as the message words it; any other project, nil.
func CheckProjectChange(project *Project, action string) error {
	if !isHivepaasProject(project) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrSystemProjectProtected).WithParam("Action", action)
}

// CheckProjectEnvChange refuses deleting or disabling an environment of the
// project HivePaaS runs in.
func CheckProjectEnvChange(project *Project, env *ProjectEnv, action string) error {
	if !isHivepaasProject(project) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrSystemProjectEnvProtected).
		WithParam("Name", env.Name).WithParam("Action", action)
}

// CheckAppDeletion refuses deleting any app of the project HivePaaS runs in:
// the stack's own run this installation, and those it provisions go with
// their feature, turned off under System.
func CheckAppDeletion(project *Project, app *App) error {
	if !isHivepaasProject(project) {
		return nil
	}
	err := hperrors.Wrap(hperrors.ErrSystemAppProtected).
		WithParam("Name", app.Name).WithParam("Action", "deleted")
	if !slices.Contains(hivepaasStackAppKeys, app.Key) {
		err = err.WithExtraDetail("It goes when its feature is turned off, on its page under System.")
	}
	return err
}

// CheckAppStop refuses stopping or disabling an app of the stack HivePaaS is
// deployed as - action says which. One HivePaaS provisions may be: its
// feature's page under System says what it is for.
func CheckAppStop(project *Project, app *App, action string) error {
	if !isHivepaasProject(project) || !slices.Contains(hivepaasStackAppKeys, app.Key) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrSystemAppProtected).WithParam("Name", app.Name).WithParam("Action", action)
}
