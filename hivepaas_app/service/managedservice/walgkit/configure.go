package walgkit

import (
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// Configure makes the engine's setting of the app and its jobs: the nightly
// base backup, the archive's health check, the restore test, and the dump when
// a repository is chosen. source is the app's deployment as the engine set its
// server's command.
func Configure(p *Profile, req *managedmodel.ConfigureReq, options managedmodel.Options,
	source *entity.AppDeploymentSettings) (*managedmodel.ConfigureResp, error) {
	managedID := gofn.Must(ulid.NewStringULID())
	restoring := req.Prepared.Restore != nil
	job := func(jobType base.SchedJobType, name, script string, schedule *entity.SchedJobSchedule,
		timeout time.Duration, env ...*entity.EnvVar) *entity.Setting {
		return NewJob(req.App.ID, managedID, name, &entity.SchedJob{
			JobType:  jobType,
			Schedule: schedule,
			Timeout:  timeutil.Duration(timeout),
			Command:  &entity.CommandTemplate{Command: ScriptDir + "/" + script, EnvVars: env},
		}, restoring, req.Now)
	}
	// The nightly run is two steps, its retention decided by the engine between
	// them (see Retention).
	nightly := job(base.SchedJobTypeManagedNightly, "Nightly base backup", "walg-nightly",
		&entity.SchedJobSchedule{CronExpr: options.BaseBackupSchedule, InitialTime: req.Now}, NightlyTimeout,
		&entity.EnvVar{Key: EnvRetainFull, Value: strconv.Itoa(options.RetainFull)})
	health := job(base.SchedJobTypeContainerCommand, "Archive health", "walg-health",
		&entity.SchedJobSchedule{Interval: timeutil.Duration(HealthInterval), InitialTime: req.Now}, HealthTimeout)
	restoreTest := job(base.SchedJobTypeContainerCommand, "Restore test", "walg-restore-test",
		&entity.SchedJobSchedule{CronExpr: options.RestoreTestSchedule, InitialTime: req.Now}, RestoreTestTimeout,
		&entity.EnvVar{Key: EnvRestoreCheckQuery, Value: options.RestoreTestQuery})
	jobs := []*entity.Setting{nightly, health, restoreTest}
	dump := entity.ManagedDump{Repo: entity.ObjectID{ID: options.DumpRepo}, Schedule: options.DumpSchedule}
	var dumpJob *entity.Setting
	if dump.Repo.ID != "" {
		dumpJob = NewDumpJob(p, req.App, managedID, dump, restoring, req.Now)
		jobs = append(jobs, dumpJob)
	}

	managed := NewAppSetting(req.App.ID, base.SettingTypeAppManaged, "", entity.CurrentAppManagedVersion,
		req.Now, &entity.AppManagedSettings{
			Engine:    string(p.Engine),
			EngineAPI: p.API,
			Locks:     Locks(req, source),
			Database: &entity.ManagedDatabase{
				Storage:        req.Storage,
				EncryptionKey:  entity.NewEncryptedField(req.Prepared.EncryptionKey),
				ArchiveTimeout: timeutil.Duration(options.ArchiveTimeout),
				RetainFull:     options.RetainFull,
				RestoreTest: entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer,
					CheckQuery: options.RestoreTestQuery},
				Dump: dump,
				Jobs: entity.ManagedJobs{
					Nightly:     entity.ObjectID{ID: nightly.ID},
					Health:      entity.ObjectID{ID: health.ID},
					RestoreTest: entity.ObjectID{ID: restoreTest.ID},
					Dump:        SettingID(dumpJob),
				},
			},
		})
	managed.ID = managedID
	if p.WriterAlias {
		data := managed.MustAsAppManagedSettings()
		data.Database.Alias = WriterAlias(req.App)
		data.Locks.Aliases = []string{data.Database.Alias}
		managed.MustSetData(data)
	}
	if restore := req.Prepared.Restore; restore != nil {
		data := managed.MustAsAppManagedSettings()
		data.Database.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: restore.SourceAppID},
			Target: restore.Target}
		managed.MustSetData(data)
	}

	return &managedmodel.ConfigureResp{Settings: append([]*entity.Setting{managed}, jobs...)}, nil
}

// SetServerCommand gives the app's server its command, on the app-deployment
// setting the document was built into: a template may not set a command, and
// this one is the engine's, re-applied by every deploy as the setting is.
// Locked fields keep a person from changing it. Every database starts through
// walg-entrypoint, which hands over to the image's.
func SetServerCommand(p *Profile, settings []*entity.Setting, command string) (*entity.AppDeploymentSettings, error) {
	var deployment *entity.Setting
	for _, setting := range settings {
		if setting.Type == base.SettingTypeAppDeployment {
			deployment = setting
		}
	}
	if deployment == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: the app deploys nothing", p.Engine)
	}
	source, err := deployment.AsAppDeploymentSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	source.Command = command
	source.Entrypoint = ScriptDir + "/walg-entrypoint"
	return source, hperrors.Wrap(deployment.SetData(source))
}

// NewAppSetting is a setting of the app's, active, made now.
func NewAppSetting(appID string, typ base.SettingType, name string, version int, now time.Time,
	data entity.SettingData) *entity.Setting {
	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  appID,
		Type:      typ,
		Name:      name,
		Status:    base.SettingStatusActive,
		Version:   version,
		CreatedAt: now,
		UpdatedAt: now,
	}
	setting.MustSetData(data)
	return setting
}

// NewJob is one of the engine's jobs on the app: managed by its setting, and
// notifying when it fails. A restored database's jobs wait for it to recover:
// until then there is no server to check, back up, test or dump, and each run
// would fail.
func NewJob(appID, managedID, name string, job *entity.SchedJob, restoring bool, now time.Time) *entity.Setting {
	job.App = entity.ObjectID{ID: appID}
	job.ManagedBy = entity.ObjectID{ID: managedID}
	job.Notification = &entity.BaseEventNotification{FailureUseDefault: true}
	setting := NewAppSetting(appID, base.SettingTypeSchedJob, name, entity.CurrentSchedJobVersion, now, job)
	setting.Kind = string(job.JobType)
	if restoring {
		setting.Status = base.SettingStatusDisabled
	}
	return setting
}

// SettingID is a setting's reference, empty for none.
func SettingID(setting *entity.Setting) entity.ObjectID {
	if setting == nil {
		return entity.ObjectID{}
	}
	return entity.ObjectID{ID: setting.ID}
}

// WithDefaults are the options chosen, the engines' defaults where none was.
func WithDefaults(options *managedmodel.Options) managedmodel.Options {
	out := managedmodel.Options{}
	if options != nil {
		out = *options
	}
	out.ArchiveTimeout = gofn.Coalesce(out.ArchiveTimeout, DefaultArchiveTimeout)
	out.RetainFull = gofn.Coalesce(out.RetainFull, DefaultRetainFull)
	out.BaseBackupSchedule = gofn.Coalesce(out.BaseBackupSchedule, DefaultBaseBackupSchedule)
	out.RestoreTestSchedule = gofn.Coalesce(out.RestoreTestSchedule, DefaultRestoreTestSchedule)
	out.DumpSchedule = gofn.Coalesce(out.DumpSchedule, DefaultDumpSchedule)
	return out
}

// Locks are what the engine owns of the app: its image's line, its command, the
// mounts and variables it relies on, and the secrets, config files and setting
// mounts the app was built with.
func Locks(req *managedmodel.ConfigureReq, source *entity.AppDeploymentSettings) entity.ManagedLocks {
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
