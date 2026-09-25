package settingmountserviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	sms "github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

// engine is a fixture over a fake Docker whose service already has an ordinary
// secret, db_password.
func engine(t *testing.T, entries []*entity.Setting, sources ...*entity.Setting) (*service, *fakeDocker) {
	t.Helper()
	svc := fixture(t, entries, sources...)
	fake := newFakeDocker(swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Secrets: []*swarm.SecretReference{{SecretID: "ordinary_1", SecretName: "shop_prod_api_db_password",
			File: &swarm.SecretReferenceFileTarget{Name: "db_password", UID: "0", GID: "0", Mode: 0o444}}},
	}}})
	svc.dockerManager = fake
	svc.logger = logging.GlobalLogger()
	return svc, fake
}

func activeCertEntry(t *testing.T) *entity.Setting {
	t.Helper()
	return entry(t, "tls-cert", base.SettingStatusActive, certFiles("cert_1"))
}

func mountedTargets(spec *swarm.ServiceSpec) (secrets, configs []string) {
	for _, ref := range spec.TaskTemplate.ContainerSpec.Secrets {
		secrets = append(secrets, ref.File.Name)
	}
	for _, ref := range spec.TaskTemplate.ContainerSpec.Configs {
		configs = append(configs, ref.File.Name)
	}
	return secrets, configs
}

func TestApplyCreatesLabeledObjectsAndReferencesThem(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))
	spec := fake.copyService().Spec

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	secrets, configs := mountedTargets(&spec)
	assert.Equal(t, []string{"db_password", "/etc/app/tls/key.pem"}, secrets, "the ordinary secret stays")
	assert.Equal(t, []string{"/etc/app/tls/cert.pem"}, configs)
	key := spec.TaskTemplate.ContainerSpec.Secrets[1]
	assert.Equal(t, "1000", key.File.UID)
	assert.EqualValues(t, 0o400, key.File.Mode)
	stored := fake.secrets[key.SecretID]
	assert.Equal(t, []byte("KEY"), stored.Spec.Data)
	assert.Equal(t, sms.Labels("app_1", "tls-cert", "privateKey"), stored.Spec.Labels)
	assert.Regexp(t, `^shop_prod_api_mount_tls-cert_privatekey_[0-9a-f]{8}$`, key.SecretName)
}

func TestTheSameDataTwiceCreatesNothingNew(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))

	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))
	created := len(fake.created)
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.Len(t, fake.created, created)
	assert.Equal(t, 1, fake.updates, "nothing changed, so the service is not updated, nor restarted")
}

// A certificate and its key are replaced in one update, and what they replace
// is removed once nothing references it.
func TestARenewedCertificateIsSwappedInOneUpdateAndTheOldOneRemoved(t *testing.T) {
	source := certSource(t, "cert_1", "CERT", "KEY")
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, source)
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.NoError(t, source.SetData(&entity.SSLCert{Certificate: "CERT2",
		PrivateKey: entity.NewEncryptedField("KEY2")}))
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	assert.Equal(t, 2, fake.updates)
	assert.Len(t, fake.secrets, 1, "the old key is gone; the ordinary secret was never Docker's to list here")
	assert.Len(t, fake.configs, 1)
	for _, secret := range fake.secrets {
		assert.Equal(t, []byte("KEY2"), secret.Spec.Data)
	}
}

func TestAnOrdinarySecretKeepsItsTarget(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"},
		Files:  []*entity.AppSettingMountFile{{Part: "privateKey", Path: "/run/secrets/db_password"}},
	})}, certSource(t, "cert_1", "CERT", "KEY"))
	spec := fake.copyService().Spec

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	assert.Len(t, spec.TaskTemplate.ContainerSpec.Secrets, 1)
	assert.Equal(t, "ordinary_1", spec.TaskTemplate.ContainerSpec.Secrets[0].SecretID)
	assert.Empty(t, fake.created)
}

// A disabled entry's files leave the service, and their objects leave Docker.
func TestAnEntryThatCannotBeUsedIsTakenOut(t *testing.T) {
	e := activeCertEntry(t)
	svc, fake := engine(t, []*entity.Setting{e}, certSource(t, "cert_1", "CERT", "KEY"))
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	e.Status = base.SettingStatusDisabled
	assert.NoError(t, svc.Refresh(context.Background(), nil, testApp))

	secrets, configs := mountedTargets(&fake.service.Spec)
	assert.Equal(t, []string{"db_password"}, secrets)
	assert.Empty(t, configs)
	assert.Empty(t, fake.secrets)
	assert.Empty(t, fake.configs)
}

// A preview or a clone starts from another app's spec; the files that app
// mounts are its own, and this app resolves its own.
func TestReferencesToAnotherAppsMountedObjectsAreDropped(t *testing.T) {
	svc, fake := engine(t, nil)
	fake.secrets["parent_key"] = swarm.Secret{ID: "parent_key", Spec: swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: "parent_key", Labels: sms.Labels("app_parent", "tls-cert", "privateKey")}}}
	spec := fake.copyService().Spec
	spec.TaskTemplate.ContainerSpec.Secrets = append(spec.TaskTemplate.ContainerSpec.Secrets,
		&swarm.SecretReference{SecretID: "parent_key", SecretName: "parent_key",
			File: &swarm.SecretReferenceFileTarget{Name: "/etc/app/tls/key.pem"}})

	assert.NoError(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))

	secrets, _ := mountedTargets(&spec)
	assert.Equal(t, []string{"db_password"}, secrets)
	assert.NoError(t, svc.Sweep(context.Background(), testApp))
	assert.Contains(t, fake.secrets, "parent_key", "another app's object is its own to sweep")
}

func TestSweepGivesUpOnWhatStaysInUseAndRemovesTheRest(t *testing.T) {
	svc, fake := engine(t, nil)
	for _, id := range []string{"stuck", "free"} {
		fake.secrets[id] = swarm.Secret{ID: id, Spec: swarm.SecretSpec{
			Annotations: swarm.Annotations{Name: id, Labels: sms.Labels("app_1", "a", "privateKey")}}}
	}
	fake.inUse["stuck"] = 100

	err := svc.Sweep(context.Background(), testApp)

	assert.Error(t, err)
	assert.Contains(t, fake.secrets, "stuck")
	assert.NotContains(t, fake.secrets, "free")
}

func TestASourceThatFailsToRenderChangesNothing(t *testing.T) {
	svc, fake := engine(t, []*entity.Setting{activeCertEntry(t)}, certSource(t, "cert_1", "CERT", "KEY"))
	svc.loadSources = func(context.Context, database.IDB, *entity.App, []string) ([]*entity.Setting, error) {
		return nil, errors.New("no data key")
	}
	spec := fake.copyService().Spec
	before := fake.copyService().Spec

	assert.Error(t, svc.ApplyToService(context.Background(), nil, testApp, &spec))
	assert.Equal(t, before, spec)
	assert.Empty(t, fake.created)
}

func TestRemoveAppRemovesOnlyItsOwnMountedObjects(t *testing.T) {
	svc, fake := engine(t, nil)
	fake.secrets["mine"] = swarm.Secret{ID: "mine", Spec: swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: "mine", Labels: sms.Labels("app_1", "a", "privateKey")}}}
	fake.configs["theirs"] = swarm.Config{ID: "theirs", Spec: swarm.ConfigSpec{
		Annotations: swarm.Annotations{Name: "theirs", Labels: sms.Labels("app_2", "a", "certificate")}}}

	assert.NoError(t, svc.RemoveApp(context.Background(), "app_1"))

	assert.Empty(t, fake.secrets)
	assert.Contains(t, fake.configs, "theirs")
}

// A service that cannot be found says nothing of what a service will need: it
// may be between removal and creation. Deleting an app is RemoveApp's.
func TestSweepRemovesNothingWhileTheServiceCannotBeFound(t *testing.T) {
	svc, fake := engine(t, nil)
	fake.secrets["needed"] = swarm.Secret{ID: "needed", Spec: swarm.SecretSpec{
		Annotations: swarm.Annotations{Name: "needed", Labels: sms.Labels("app_1", "a", "privateKey")}}}
	fake.gone = true

	assert.NoError(t, svc.Sweep(context.Background(), testApp))
	noService := *testApp
	noService.ServiceID = ""
	assert.NoError(t, svc.Sweep(context.Background(), &noService))

	assert.Contains(t, fake.secrets, "needed")
}
