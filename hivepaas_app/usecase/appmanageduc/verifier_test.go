package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

func sourceSettings(t *testing.T) []*entity.Setting {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	env := setting("env-1", base.SettingTypeEnvVar, &entity.EnvVars{Data: []*entity.EnvVar{
		{Key: "WALG_S3_PREFIX", Value: "s3://b/walg/db"}}})
	access := secretSetting(t, "sec-1", "WALG_AWS_ACCESS_KEY_ID", "AKIA-DB")
	sodium := secretSetting(t, "sec-2", "WALG_LIBSODIUM_KEY", "sodium")
	script := setting("cfg-1", base.SettingTypeConfigFile, &entity.ConfigFile{Name: "walg-run"})
	scriptMount := setting("mnt-1", base.SettingTypeAppSettingMount, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cfg-1"}, Files: []*entity.AppSettingMountFile{{Path: "/opt/hivepaas/bin/walg-run"}}})
	secretMount := setting("mnt-2", base.SettingTypeAppSettingMount, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "sec-1"}, Files: []*entity.AppSettingMountFile{{Path: "/run/secrets/walg/x"}}})
	deployment := setting("dep-1", base.SettingTypeAppDeployment, &entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage, Command: "postgres -c archive_mode=on",
		PreDeploymentCommand: "echo before", PostDeploymentCommand: "echo after",
		ImageSource: &entity.DeploymentImageSource{Image: "postgres:17.6-trixie"}})
	projectMount := setting("mnt-3", base.SettingTypeAppSettingMount, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "project-cfg"}, Files: []*entity.AppSettingMountFile{{Path: "/etc/shared"}}})
	artifacts := setting("art-1", base.SettingTypeAppArtifacts, &entity.AppArtifactsSettings{})
	routing := setting("rt-1", base.SettingTypeAppRouting, &entity.AppRoutingSettings{})
	all := []*entity.Setting{env, access, sodium, script, scriptMount, secretMount, projectMount, deployment,
		artifacts, routing}
	for _, s := range all {
		s.Scope, s.ObjectID = base.ObjectScopeApp, "db-1"
	}
	return all
}

// The verifier is the database's image, artifacts, scripts, variables and
// credentials, idling: the restore test starts the server it needs itself.
func TestTheVerifierIsTheDatabaseIdling(t *testing.T) {
	now := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)

	settings, err := verifierSettings(sourceSettings(t), &entity.App{ID: "ver-1"}, now,
		map[string]string{"WALG_AWS_ACCESS_KEY_ID": "AKIA-READONLY"},
		managedmodel.VerifierSpec{Command: "sleep infinity"})

	assert.NoError(t, err)
	byID := map[string]*entity.Setting{}
	types := map[base.SettingType]int{}
	for _, s := range settings {
		assert.Equal(t, "ver-1", s.ObjectID, s.Type)
		assert.Equal(t, base.ObjectScopeApp, s.Scope)
		assert.NotContains(t, []string{"env-1", "sec-1", "sec-2", "cfg-1", "mnt-1", "mnt-2", "mnt-3", "dep-1",
			"art-1"}, s.ID,
			"a setting of its own")
		byID[s.ID] = s
		types[s.Type]++
	}
	assert.Equal(t, map[base.SettingType]int{base.SettingTypeEnvVar: 1, base.SettingTypeSecret: 2,
		base.SettingTypeConfigFile: 1, base.SettingTypeAppSettingMount: 3, base.SettingTypeAppDeployment: 1,
		base.SettingTypeAppArtifacts: 1}, types, "not its routing: it serves nothing")

	for _, s := range settings {
		switch s.Type { //nolint:exhaustive // the types the verifier is given
		case base.SettingTypeAppDeployment:
			assert.Equal(t, "sleep infinity", s.MustAsAppDeploymentSettings().Command)
			assert.Empty(t, s.MustAsAppDeploymentSettings().PreDeploymentCommand, "no task to run it in")
			assert.Empty(t, s.MustAsAppDeploymentSettings().PostDeploymentCommand, "no task to run it in")
			assert.Equal(t, "postgres:17.6-trixie", s.MustAsAppDeploymentSettings().ImageSource.Image)
		case base.SettingTypeSecret:
			secret, _ := s.AsSecret()
			plain, _ := secret.Value.GetPlain()
			if secret.Key == "WALG_AWS_ACCESS_KEY_ID" {
				assert.Equal(t, "AKIA-READONLY", plain, "the narrower key")
			} else {
				assert.Equal(t, "sodium", plain)
			}
		case base.SettingTypeAppSettingMount:
			source := s.MustAsAppSettingMount().Source.ID
			if source == "project-cfg" {
				continue // the project's, shared
			}
			if assert.Contains(t, byID, source, "a mount of the verifier's own setting") {
				assert.Contains(t, []base.SettingType{base.SettingTypeConfigFile, base.SettingTypeSecret},
					byID[source].Type)
			}
		}
	}
}
