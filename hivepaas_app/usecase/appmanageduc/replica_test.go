package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// replicaSources are a primary's settings: variables, a secret and the mount of
// it, a script and its mount, the deployment, and its routing, which a replica is
// not given.
func replicaSources() []*entity.Setting {
	secret := setting("secret-1", base.SettingTypeSecret, &entity.Secret{Key: "WALG_LIBSODIUM_KEY",
		Value: entity.NewEncryptedField("sodium")})
	secret.Name = "WALG_LIBSODIUM_KEY"
	return []*entity.Setting{
		setting("env-1", base.SettingTypeEnvVar, &entity.EnvVars{Data: []*entity.EnvVar{
			{Key: "POSTGRES_DB", Value: "${HIVEPAAS_DATABASE_NAME}"},
			{Key: "WALG_S3_PREFIX", Value: "s3://backups/walg/shop-db", IsLiteral: true}}}),
		secret,
		setting("mount-1", base.SettingTypeAppSettingMount, &entity.AppSettingMount{
			Source: entity.ObjectID{ID: "secret-1"},
			Files:  []*entity.AppSettingMountFile{{Part: "value", Path: "/run/secrets/walg/WALG_LIBSODIUM_KEY"}}}),
		setting("script-1", base.SettingTypeConfigFile, &entity.ConfigFile{}),
		setting("mount-2", base.SettingTypeAppSettingMount, &entity.AppSettingMount{
			Source: entity.ObjectID{ID: "script-1"},
			Files:  []*entity.AppSettingMountFile{{Part: "content", Path: "/opt/hivepaas/bin/walg-replication"}}}),
		setting("deploy-1", base.SettingTypeAppDeployment, &entity.AppDeploymentSettings{
			ActiveMethod: base.DeploymentMethodImage, Command: "postgres -c archive_mode=on",
			PreDeploymentCommand: "migrate", RestorePointBeforeDeploy: entity.ObjectID{ID: "db-2"}}),
		setting("routing-1", base.SettingTypeAppRouting, &entity.AppRoutingSettings{}),
	}
}

func replicaSpec() *managedmodel.ReplicaSpec {
	return &managedmodel.ReplicaSpec{
		Env: [][2]string{{"HIVEPAAS_STANDBY_HOST", "shop-db"}, {"HIVEPAAS_STANDBY_NAME", "shop-db-r1"},
			{"HIVEPAAS_STANDBY_PASSWORD_FILE", "/run/secrets/hivepaas-replication/password"}},
		SecretName:   "HIVEPAAS_REPLICATION_PASSWORD",
		PasswordFile: "/run/secrets/hivepaas-replication/password",
		DataTarget:   "/var/lib/postgresql",
	}
}

func byType(settings []*entity.Setting, typ base.SettingType) []*entity.Setting {
	var out []*entity.Setting
	for _, one := range settings {
		if one.Type == typ {
			out = append(out, one)
		}
	}
	return out
}

// A replica is given the primary's settings as its own, the password as a
// secret file, the standby's variables, and a setting naming its primary with
// its own locks.
func TestAReplicasSettings(t *testing.T) {
	primary := loadedPostgres(t)
	primary.app = &entity.App{ID: "app-1", Key: "shop-db", Name: "shop-db"}
	primary.managed.Locks = entity.ManagedLocks{Image: "postgres:18.6-trixie", ImageLine: "18",
		Command: "postgres -c archive_mode=on", Mounts: []string{"/var/lib/postgresql"}, EnvKeys: []string{"PGDATA"},
		Settings: []string{"secret-1", "mount-1", "script-1", "mount-2"}}
	replica := &entity.App{ID: "replica-1", Key: "shop-db-r1"}

	settings, err := replicaSettings(replicaSources(), replica, primary, replicaSpec(), "ReplPass0123456789",
		time.Now().UTC())

	assert.NoError(t, err)
	assert.Empty(t, byType(settings, base.SettingTypeAppRouting), "it serves nothing of its own")
	for _, one := range settings {
		assert.Equal(t, "replica-1", one.ObjectID)
		assert.NotContains(t, []string{"env-1", "secret-1", "mount-1", "script-1", "mount-2", "deploy-1"}, one.ID,
			"a copy has an id of its own")
	}

	vars := byType(settings, base.SettingTypeEnvVar)[0].MustAsEnvVars().Data
	byKey := map[string]*entity.EnvVar{}
	for _, env := range vars {
		byKey[env.Key] = env
	}
	assert.Equal(t, "s3://backups/walg/shop-db", byKey["WALG_S3_PREFIX"].Value, "the primary's archive")
	assert.Equal(t, "shop-db", byKey["HIVEPAAS_STANDBY_HOST"].Value)
	assert.True(t, byKey["HIVEPAAS_STANDBY_HOST"].IsLiteral)
	assert.Equal(t, "shop-db-r1", byKey["HIVEPAAS_STANDBY_NAME"].Value)
	for _, env := range vars {
		assert.NotContains(t, env.Value, "ReplPass", "the password is in no variable")
	}

	deployment := byType(settings, base.SettingTypeAppDeployment)[0].MustAsAppDeploymentSettings()
	assert.Equal(t, "postgres -c archive_mode=on", deployment.Command, "the server's command stays")
	assert.Empty(t, deployment.PreDeploymentCommand)
	assert.Empty(t, deployment.RestorePointBeforeDeploy.ID)

	var password *entity.Setting
	for _, one := range byType(settings, base.SettingTypeSecret) {
		if one.Name == "HIVEPAAS_REPLICATION_PASSWORD" {
			password = one
		}
	}
	if !assert.NotNil(t, password) {
		return
	}
	value, _ := password.MustAsSecret().Value.GetPlain()
	assert.Equal(t, "ReplPass0123456789", value)
	var passwordFile *entity.AppSettingMountFile
	for _, one := range byType(settings, base.SettingTypeAppSettingMount) {
		mount := one.MustAsAppSettingMount()
		if mount.Source.ID == password.ID {
			passwordFile = mount.Files[0]
		}
		assert.NotContains(t, []string{"secret-1", "script-1"}, mount.Source.ID, "mounts of the copies")
	}
	if assert.NotNil(t, passwordFile) {
		assert.Equal(t, "/run/secrets/hivepaas-replication/password", passwordFile.Path)
		assert.Equal(t, "999", passwordFile.UID)
		assert.Equal(t, "0400", passwordFile.Mode.String())
	}

	managed := byType(settings, base.SettingTypeAppManaged)
	if !assert.Len(t, managed, 1) {
		return
	}
	member := managed[0].MustAsAppManagedSettings()
	assert.Equal(t, &entity.ManagedMemberOf{App: entity.ObjectID{ID: "app-1"}, Role: entity.ManagedRoleReplica},
		member.MemberOf)
	assert.Nil(t, member.Database, "its engine state is its primary's")
	assert.Equal(t, "postgres:18.6-trixie", member.Locks.Image)
	assert.Equal(t, []string{"/var/lib/postgresql"}, member.Locks.Mounts)
	assert.Contains(t, member.Locks.EnvKeys, "HIVEPAAS_STANDBY_HOST")
	assert.Len(t, member.Locks.Settings, 6, "the four copies, the password and its mount")
	assert.Contains(t, member.Locks.Settings, password.ID)
}

func TestAReplicasKey(t *testing.T) {
	pg := &entity.ManagedDatabase{}
	assert.Equal(t, "shop-db-r1", replicaKey("shop-db", pg))

	pg.Members = []*entity.ManagedMember{{Key: "shop-db-r1"}, {Key: "shop-db-r3"}}
	assert.Equal(t, "shop-db-r2", replicaKey("shop-db", pg), "the first one free")
}

func TestAReplicationPassword(t *testing.T) {
	one, other := newReplicationPassword(), newReplicationPassword()
	assert.Regexp(t, `^[A-Za-z0-9]{40}$`, one)
	assert.NotEqual(t, one, other)
}

// A replica is removed from its primary, not unmanaged; a primary loses its
// replicas before it is unmanaged.
func TestMembersAndTheirPrimaryAreNotUnmanaged(t *testing.T) {
	member := &entity.App{ID: "r1", Name: "shop-db r1", Settings: []*entity.Setting{
		setting("managed-r", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres",
			MemberOf: &entity.ManagedMemberOf{App: entity.ObjectID{ID: "app-1"}, Role: entity.ManagedRoleReplica}}),
	}}
	_, err := unmanagedSettings(member, time.Now())
	assert.ErrorIs(t, err, hperrors.ErrAppManagedIsMember)

	primary := &entity.App{ID: "app-1", Name: "shop-db", Settings: []*entity.Setting{
		setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres",
			Database: &entity.ManagedDatabase{Members: []*entity.ManagedMember{{App: entity.ObjectID{ID: "r1"}}}}}),
	}}
	_, err = unmanagedSettings(primary, time.Now())
	assert.ErrorIs(t, err, hperrors.ErrAppManagedHasMembers)
}
