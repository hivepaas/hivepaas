package appmanageduc

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// verifierSourceTypes are the database's settings a verifier is given: what it
// needs to restore the archive the database writes - its image, artifacts,
// scripts, variables and credentials, and its kind, whose database user and
// password the variables refer to (POSTGRES_USER is ${HIVEPAAS_USER}). Not its
// routing: it serves nothing.
var verifierSourceTypes = []base.SettingType{
	base.SettingTypeEnvVar, base.SettingTypeSecret, base.SettingTypeConfigFile,
	base.SettingTypeAppSettingMount, base.SettingTypeAppDeployment, base.SettingTypeAppArtifacts,
	base.SettingTypeAppKind,
}

// verifierSettings are the database's settings, the verifier's own: new ids,
// mounts of the copies rather than of the database's, the command the engine
// idles with and no deployment commands, and secret values in place of the
// database's where given - a narrower key, read-only where the storage has one.
func verifierSettings(
	sources []*entity.Setting,
	verifier *entity.App,
	now time.Time,
	secretValues map[string]string,
	spec managedmodel.VerifierSpec,
) ([]*entity.Setting, error) {
	ids := make(map[string]string, len(sources))
	out := make([]*entity.Setting, 0, len(sources))
	for _, source := range sources {
		if !gofn.Contain(verifierSourceTypes, source.Type) {
			continue
		}
		setting := &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       base.ObjectScopeApp,
			ObjectID:    verifier.ID,
			Type:        source.Type,
			Kind:        source.Kind,
			Name:        source.Name,
			Status:      base.SettingStatusActive,
			Inheritable: source.Inheritable,
			Version:     source.Version,
			Data:        source.Data,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		ids[source.ID] = setting.ID
		out = append(out, setting)
	}
	for _, setting := range out {
		if err := adaptForVerifier(setting, ids, secretValues, spec); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func adaptForVerifier(setting *entity.Setting, ids, secretValues map[string]string,
	spec managedmodel.VerifierSpec) error {
	switch setting.Type { //nolint:exhaustive // three of verifierSourceTypes change
	case base.SettingTypeAppSettingMount:
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			return hperrors.Wrap(err)
		}
		// A source of the project's or the env's is shared, and stays.
		if id, found := ids[mount.Source.ID]; found {
			mount.Source = entity.ObjectID{ID: id}
		}
		return hperrors.Wrap(setting.SetData(mount))
	case base.SettingTypeAppDeployment:
		deployment, err := setting.AsAppDeploymentSettings()
		if err != nil {
			return hperrors.Wrap(err)
		}
		// Nothing runs around a deployment: at none, there is no task to run in.
		deployment.Command = spec.Command
		deployment.PreDeploymentCommand, deployment.PostDeploymentCommand = "", ""
		deployment.RestorePointBeforeDeploy = entity.ObjectID{}
		return hperrors.Wrap(setting.SetData(deployment))
	case base.SettingTypeSecret:
		value, found := secretValues[setting.Name]
		if !found {
			return nil
		}
		secret, err := setting.AsSecret()
		if err != nil {
			return hperrors.Wrap(err)
		}
		secret.Value = entity.NewEncryptedField(value)
		return hperrors.Wrap(setting.SetData(secret))
	}
	return nil
}
