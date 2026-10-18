package postgres

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	defaultArchiveTimeout      = time.Minute
	defaultRetainFull          = 7
	defaultBaseBackupSchedule  = "0 2 * * *"
	defaultRestoreTestSchedule = "0 4 * * 0"
	healthInterval             = 5 * time.Minute

	nightlyTimeout     = 6 * time.Hour
	healthTimeout      = 2 * time.Minute
	restoreTestTimeout = 6 * time.Hour

	// sharedBuffersShare is the memory limit's share Postgres caches pages in,
	// the usual quarter, and minSharedBuffers its floor.
	sharedBuffersShare = 4
	minSharedBuffers   = 128 * unit.MB
)

func (e *Engine) Configure(req *managedmodel.ConfigureReq) (*managedmodel.ConfigureResp, error) {
	options := withDefaults(req.Prepared.Options)
	source, err := setServerArguments(req.Settings, req.Prepared.Doc, options)
	if err != nil {
		return nil, err
	}

	newSetting := func(typ base.SettingType, name string, version int, data entity.SettingData) *entity.Setting {
		setting := &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  req.App.ID,
			Type:      typ,
			Name:      name,
			Status:    base.SettingStatusActive,
			Version:   version,
			CreatedAt: req.Now,
			UpdatedAt: req.Now,
		}
		setting.MustSetData(data)
		return setting
	}

	managedID := gofn.Must(ulid.NewStringULID())
	job := func(name, script string, schedule *entity.SchedJobSchedule, timeout time.Duration) *entity.Setting {
		setting := newSetting(base.SettingTypeSchedJob, name, entity.CurrentSchedJobVersion, &entity.SchedJob{
			JobType:      base.SchedJobTypeContainerCommand,
			Schedule:     schedule,
			App:          entity.ObjectID{ID: req.App.ID},
			Timeout:      timeutil.Duration(timeout),
			Command:      &entity.CommandTemplate{Command: scriptDir + "/" + script},
			Notification: &entity.BaseEventNotification{FailureUseDefault: true},
			ManagedBy:    entity.ObjectID{ID: managedID},
		})
		setting.Kind = string(base.SchedJobTypeContainerCommand)
		return setting
	}
	nightly := job("Nightly base backup", "walg-nightly",
		&entity.SchedJobSchedule{CronExpr: options.BaseBackupSchedule, InitialTime: req.Now}, nightlyTimeout)
	health := job("Archive health", "walg-health",
		&entity.SchedJobSchedule{Interval: timeutil.Duration(healthInterval), InitialTime: req.Now}, healthTimeout)
	restoreTest := job("Restore test", "walg-restore-test",
		&entity.SchedJobSchedule{CronExpr: options.RestoreTestSchedule, InitialTime: req.Now}, restoreTestTimeout)

	managed := newSetting(base.SettingTypeAppManaged, "", entity.CurrentAppManagedVersion, &entity.AppManagedSettings{
		Engine:    string(e.Name()),
		EngineAPI: maxAPI,
		Locks:     locks(req, source),
		Postgres: &entity.ManagedPostgres{
			Storage:        req.Storage,
			EncryptionKey:  entity.NewEncryptedField(req.Prepared.EncryptionKey),
			ArchiveTimeout: timeutil.Duration(options.ArchiveTimeout),
			RetainFull:     options.RetainFull,
			RestoreTest: entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer,
				CheckQuery: options.RestoreTestQuery},
			Jobs: entity.ManagedPostgresJobs{
				Nightly:     entity.ObjectID{ID: nightly.ID},
				Health:      entity.ObjectID{ID: health.ID},
				RestoreTest: entity.ObjectID{ID: restoreTest.ID},
			},
		},
	})
	managed.ID = managedID

	return &managedmodel.ConfigureResp{Settings: []*entity.Setting{managed, nightly, health, restoreTest}}, nil
}

func withDefaults(options *managedmodel.Options) managedmodel.Options {
	out := managedmodel.Options{}
	if options != nil {
		out = *options
	}
	out.ArchiveTimeout = gofn.Coalesce(out.ArchiveTimeout, defaultArchiveTimeout)
	out.RetainFull = gofn.Coalesce(out.RetainFull, defaultRetainFull)
	out.BaseBackupSchedule = gofn.Coalesce(out.BaseBackupSchedule, defaultBaseBackupSchedule)
	out.RestoreTestSchedule = gofn.Coalesce(out.RestoreTestSchedule, defaultRestoreTestSchedule)
	return out
}

// setServerArguments gives the server what archiving needs, on the
// app-deployment setting the document was built into: a template may not set a
// command, and these are the engine's, re-applied by every deploy as the setting
// is. Locked fields keep a person from changing them.
func setServerArguments(
	settings []*entity.Setting,
	doc *specmodel.AppDoc,
	options managedmodel.Options,
) (*entity.AppDeploymentSettings, error) {
	var deployment *entity.Setting
	for _, setting := range settings {
		if setting.Type == base.SettingTypeAppDeployment {
			deployment = setting
		}
	}
	if deployment == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("postgres: the app deploys nothing")
	}
	source, err := deployment.AsAppDeploymentSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	args := []string{
		"postgres",
		"-c wal_level=replica",
		"-c archive_mode=on",
		fmt.Sprintf("-c archive_timeout=%ds", int(options.ArchiveTimeout.Seconds())),
		`-c "archive_command=` + scriptDir + `/walg-run wal-push %p"`,
	}
	if buffers := sharedBuffers(doc); buffers > 0 {
		args = append(args, fmt.Sprintf("-c shared_buffers=%dMB", buffers/unit.MB))
	}
	source.Command = strings.Join(args, " ")
	return source, hperrors.Wrap(deployment.SetData(source))
}

// locks are what the engine owns of the app: its image's line, its command, the
// mounts and variables it relies on, and the secrets, config files and setting
// mounts the app was built with.
func locks(req *managedmodel.ConfigureReq, source *entity.AppDeploymentSettings) entity.ManagedLocks {
	out := entity.ManagedLocks{
		ImageLine: req.Prepared.Version,
		Command:   source.Command,
		EnvKeys:   slices.Clone(req.Prepared.OwnedEnv),
	}
	if source.ImageSource != nil {
		out.Image = source.ImageSource.Image
	}
	if doc := req.Prepared.Doc; doc.Deployment != nil && doc.Deployment.Storage != nil {
		out.Mounts = slices.Sorted(maps.Keys(doc.Deployment.Storage.Mounts))
	}
	for _, setting := range req.Settings {
		switch setting.Type { //nolint:exhaustive // three types hold what the engine built the app with
		case base.SettingTypeSecret, base.SettingTypeConfigFile, base.SettingTypeAppSettingMount:
			out.Settings = append(out.Settings, setting.ID)
		}
	}
	return out
}

// sharedBuffers is a quarter of the memory limit, at least minSharedBuffers;
// zero without a limit, which leaves Postgres its own default.
func sharedBuffers(doc *specmodel.AppDoc) unit.DataSize {
	if doc.Deployment == nil || doc.Deployment.Resources == nil || doc.Deployment.Resources.Limits == nil {
		return 0
	}
	limit := doc.Deployment.Resources.Limits.Memory
	if limit <= 0 {
		return 0
	}
	return max(limit/sharedBuffersShare, minSharedBuffers)
}
