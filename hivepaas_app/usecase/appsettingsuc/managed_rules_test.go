package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const lockedCommand = `postgres -c archive_mode=on -c "archive_command=/opt/hivepaas/bin/walg-run wal-push %p"`

func testLocks() *entity.ManagedLocks {
	return &entity.ManagedLocks{
		Image: "postgres:17.6-trixie", ImageLine: "17", Command: lockedCommand,
		Mounts:   []string{"/var/lib/postgresql"},
		EnvKeys:  []string{"WALG_S3_PREFIX", "PGDATA"},
		Settings: []string{"secret-1"},
	}
}

// assertLocked checks err is the refusal of a field.
func assertLocked(t *testing.T, err error, field string) {
	t.Helper()
	assert.ErrorIs(t, err, hperrors.ErrAppManagedFieldLocked)
	info, _ := hperrors.ParseError(err, "en")
	assert.Equal(t, field, info.Params["Field"])
}

func TestManagedLocksOf(t *testing.T) {
	managed := &entity.Setting{Type: base.SettingTypeAppManaged, Status: base.SettingStatusActive}
	managed.MustSetData(&entity.AppManagedSettings{Engine: "postgres", Locks: *testLocks()})

	locks, err := managedLocksOf(&entity.App{Settings: []*entity.Setting{managed}})
	assert.NoError(t, err)
	assert.Equal(t, testLocks(), locks)

	locks, err = managedLocksOf(&entity.App{})
	assert.NoError(t, err)
	assert.Nil(t, locks, "an app no engine looks after has no locks")
}

func TestAManagedAppsImageMovesWithinItsLine(t *testing.T) {
	locks := testLocks()
	for image, allowed := range map[string]bool{
		"postgres:17.6-trixie":                   true,
		"postgres:17.7-trixie":                   true,
		"docker.io/library/postgres:17.7-trixie": true,
		"postgres:18.0-trixie":                   false, // another major, another data directory
		"postgres:17.7-alpine3.22":               false, // musl: the artifacts do not run
		"postgres:17-trixie":                     false, // moving
		"bitnami/postgresql:17.6.0":              false,
	} {
		err := checkManagedImage(locks, image)
		if allowed {
			assert.NoError(t, err, image)
		} else {
			assertLocked(t, err, "image")
		}
	}
	assert.NoError(t, checkManagedImage(nil, "anything:1.0"), "nothing is locked on an ordinary app")
}

func TestAManagedAppsCommandKeepsTheEnginesArguments(t *testing.T) {
	locks := testLocks()
	for command, allowed := range map[string]bool{
		lockedCommand:                       true,
		lockedCommand + " -c work_mem=64MB": true,
		`postgres  -c archive_mode=on  -c 'archive_command=/opt/hivepaas/bin/walg-run wal-push %p'`: true,
		`postgres -c archive_mode=off`: false,
		"postgres":                     false,
		"":                             false,
	} {
		err := checkManagedCommand(locks, command)
		if allowed {
			assert.NoError(t, err, command)
		} else {
			assertLocked(t, err, "command")
		}
	}
}

func TestAManagedAppsKindStays(t *testing.T) {
	current := &entity.AppKindSettings{Category: base.AppCategoryDatabase, Engine: "postgres"}

	assert.NoError(t, checkManagedKind(testLocks(), current, base.AppCategoryDatabase, "postgres"))
	assertLocked(t, checkManagedKind(testLocks(), current, base.AppCategoryWebapp, "postgres"), "kind")
	assertLocked(t, checkManagedKind(testLocks(), current, base.AppCategoryDatabase, "mysql"), "kind")
	assert.NoError(t, checkManagedKind(nil, current, base.AppCategoryWebapp, "x"))
}

func TestAManagedAppsDeploymentKeepsTheEnginesParts(t *testing.T) {
	current := &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage,
		ImageSource: &entity.DeploymentImageSource{Image: "postgres:17.6-trixie"}, Command: lockedCommand}
	next := func(change func(s *entity.AppDeploymentSettings)) *entity.AppDeploymentSettings {
		s := *current
		image := *current.ImageSource
		s.ImageSource = &image
		change(&s)
		return &s
	}

	assert.NoError(t, checkManagedDeployment(testLocks(), current, next(func(s *entity.AppDeploymentSettings) {
		s.ImageSource.Image = "postgres:17.7-trixie"
		s.Command += " -c work_mem=64MB"
		s.PreDeploymentCommand = "echo hi"
	})))
	assertLocked(t, checkManagedDeployment(testLocks(), current, next(func(s *entity.AppDeploymentSettings) {
		s.ActiveMethod = base.DeploymentMethodRepo
	})), "deploymentMethod")
	assertLocked(t, checkManagedDeployment(testLocks(), current, next(func(s *entity.AppDeploymentSettings) {
		s.Entrypoint = "sh"
	})), "entrypoint")
	assertLocked(t, checkManagedDeployment(testLocks(), current, next(func(s *entity.AppDeploymentSettings) {
		s.Command = "postgres"
	})), "command")
}

func TestAManagedAppsContainerKeepsTheEnginesParts(t *testing.T) {
	current := &swarm.ContainerSpec{Image: "postgres:17.6-trixie", User: "",
		Args: []string{"postgres", "-c", "archive_mode=on", "-c",
			"archive_command=/opt/hivepaas/bin/walg-run wal-push %p"}}

	assert.NoError(t, checkManagedContainer(testLocks(), current, "postgres:17.7-trixie", "",
		lockedCommand+" -c work_mem=64MB", ""))
	assertLocked(t, checkManagedContainer(testLocks(), current, "postgres:17.6-trixie", "sh", lockedCommand, ""),
		"entrypoint")
	assertLocked(t, checkManagedContainer(testLocks(), current, "postgres:17.6-trixie", "", lockedCommand,
		"1000"), "user")
	assertLocked(t, checkManagedContainer(testLocks(), current, "postgres:18.0-trixie", "", lockedCommand, ""),
		"image")
}

func TestAManagedAppsMountsKeepTheirSources(t *testing.T) {
	data := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-data", Target: "/var/lib/postgresql"}
	extra := mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-extra", Target: "/extra"}
	current := []mount.Mount{data}

	assert.NoError(t, checkManagedMounts(testLocks(), current, []mount.Mount{data, extra}), "adding is allowed")
	assertLocked(t, checkManagedMounts(testLocks(), current, []mount.Mount{extra}), "storage")
	moved := data
	moved.Source = "hp-vol-other"
	assertLocked(t, checkManagedMounts(testLocks(), current, []mount.Mount{moved}), "storage")
	readOnly := data
	readOnly.ReadOnly = true
	assertLocked(t, checkManagedMounts(testLocks(), current, []mount.Mount{readOnly}), "storage")
}

func TestAManagedAppsVariablesKeepTheirValues(t *testing.T) {
	prefix := &entity.EnvVar{Key: "WALG_S3_PREFIX", Value: "s3://b/p"}
	mine := &entity.EnvVar{Key: "TZ", Value: "UTC"}
	current := []*entity.EnvVar{prefix, mine}

	assert.NoError(t, checkManagedEnv(testLocks(), current, []*entity.EnvVar{prefix,
		{Key: "TZ", Value: "Asia/Ho_Chi_Minh"}}), "the person's own may change")
	assertLocked(t, checkManagedEnv(testLocks(), current, []*entity.EnvVar{mine}), "envVars")
	assertLocked(t, checkManagedEnv(testLocks(), current, []*entity.EnvVar{mine,
		{Key: "WALG_S3_PREFIX", Value: "s3://elsewhere"}}), "envVars")
	assertLocked(t, checkManagedEnv(testLocks(), current, []*entity.EnvVar{prefix, mine,
		{Key: "PGDATA", Value: "/tmp"}}), "envVars")
}
