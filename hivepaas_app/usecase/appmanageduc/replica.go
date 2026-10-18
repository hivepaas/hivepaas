package appmanageduc

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

const (
	// replicationPasswordLen is the replication password's length, in letters and
	// digits: walg-replication takes 16 to 128.
	replicationPasswordLen = 40
	// replicaSecretMountName names the setting that mounts the password's file.
	replicaSecretMountName = "hp-replication-pass" //nolint:gosec // G101: a setting's name
	// replicaSecretMode is the password file's: the postgres user's, read-only.
	replicaSecretMode = "0400"
	postgresFileOwner = "999"
)

// replicaSourceTypes are the primary's settings a replica is given: the
// verifier's, as it runs the same image, scripts and archive.
var replicaSourceTypes = verifierSourceTypes

// replicaPlan is a replica resolved before the transaction: its volume and node,
// and its artifacts, already on the node.
type replicaPlan struct {
	volumeID       string
	node           string
	artifactMounts []mount.Mount
}

// AddManagedReplica makes an asynchronous read replica of a managed database: a
// logical child on a volume of its own, seeded from the archive and streaming
// from the database, which is told to let it in first.
func (uc *UC) AddManagedReplica(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.AddManagedReplicaReq,
) (_ *appmanageddto.AddManagedReplicaResp, err error) {
	plan, err := uc.planReplica(ctx, req)
	if err != nil {
		return nil, err
	}
	var made *appprovisionservice.ProvisionAppsResp
	committed := false
	defer func() {
		// The replica's service is created inside the transaction but is not
		// part of it.
		if err != nil && !committed && made != nil && made.Cleanup != nil {
			err = errors.Join(err, made.Cleanup(context.WithoutCancel(ctx)))
		}
	}()
	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		var txErr error
		made, txErr = uc.addReplica(ctx, db, auth, req, plan)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	committed = true
	replica := made.Apps[0]
	if err = uc.taskQueue.ScheduleTask(ctx, replica.DeploymentTask); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, replica.CertTasks...); err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &appmanageddto.AddManagedReplicaDataResp{App: &basedto.ObjectIDResp{ID: replica.App.ID}}
	if replica.DeploymentTask != nil {
		resp.Task = &basedto.ObjectIDResp{ID: replica.DeploymentTask.ID}
	}
	return &appmanageddto.AddManagedReplicaResp{Data: resp}, nil
}

// planReplica refuses what would be refused later, before anything is
// downloaded, and puts the artifacts on the replica's node.
func (uc *UC) planReplica(ctx context.Context, req *appmanageddto.AddManagedReplicaReq) (*replicaPlan, error) {
	err := uc.licenseService.CheckFeatures(ctx, []string{string(base.LicenseFeatureManagedDatabases)})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	paths, err := uc.mountPaths(ctx, uc.db, loaded.app.ID)
	if err != nil {
		return nil, err
	}
	spec, err := loaded.engine.Replica(&managedmodel.ReplicaReq{Primary: loaded.app, Managed: loaded.managed,
		MountPaths: paths})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	pin, err := uc.verifierVolumePin(ctx, appScope(loaded.app), req.DataVolume.ID, spec.DataTarget)
	if err != nil {
		return nil, err
	}
	plan := &replicaPlan{volumeID: req.DataVolume.ID, node: gofn.Coalesce(pin.NodeID, pin.NodeLabel)}
	artifacts, err := uc.artifactService.ArtifactsOf(ctx, uc.db, loaded.app.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	constraint, _ := placementservice.VolumePinConstraint([]placementservice.VolumePin{pin})
	plan.artifactMounts, err = uc.artifactService.MountsFor(ctx, artifacts, &swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{Placement: &swarm.Placement{Constraints: []string{constraint}}},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return plan, nil
}

func (uc *UC) addReplica(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appmanageddto.AddManagedReplicaReq,
	plan *replicaPlan,
) (*appprovisionservice.ProvisionAppsResp, error) {
	loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return nil, err
	}
	sources, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", loaded.app.ID),
		bunex.SelectWhereIn("setting.type IN (?)", replicaSourceTypes...),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	pg := loaded.managed.Postgres
	key := replicaKey(loaded.app.Key, pg)
	// Checked again under the lock: a restore or a move may have happened since.
	spec, err := loaded.engine.Replica(&managedmodel.ReplicaReq{Primary: loaded.app, Managed: loaded.managed,
		Key: key, MountPaths: mountPathsOf(sources)})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	password, err := pg.ReplicationPassword.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if password == "" {
		password = newReplicationPassword()
		pg.ReplicationPassword = entity.NewEncryptedField(password)
	}
	// First: a database that does not let the replica in would leave it waiting.
	if err = uc.managedService.AllowReplicas(ctx, db, loaded.app, spec.Allow, password); err != nil {
		return nil, hperrors.Wrap(err)
	}
	made, err := uc.provisionReplica(ctx, db, auth, loaded, sources, plan, spec, key, password)
	if err != nil {
		return made, err
	}
	now := timeutil.NowUTC()
	replica := made.Apps[0].App
	pg.Members = append(pg.Members, &entity.ManagedMember{App: entity.ObjectID{ID: replica.ID}, Key: replica.Key,
		Role: entity.ManagedRoleReplica, Volume: entity.ObjectID{ID: plan.volumeID}, Node: plan.node, At: now})
	if err = uc.saveManagedSetting(ctx, db, loaded.setting, loaded.managed, now); err != nil {
		return made, err
	}
	return made, uc.recordReplica(ctx, db, auth, loaded.app, "replicaAdded", replica.ID)
}

// provisionReplica makes the replica app, a logical child of the primary: its
// settings copied, its volume and artifacts mounted. It returns what it made
// even when it fails, for the caller to clean up.
func (uc *UC) provisionReplica(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	primary *loadedManaged,
	sources []*entity.Setting,
	plan *replicaPlan,
	spec *managedmodel.ReplicaSpec,
	key, password string,
) (*appprovisionservice.ProvisionAppsResp, error) {
	app := primary.app
	now := timeutil.NowUTC()
	configure := func(ctx context.Context, db database.IDB, replica *entity.App, svc *swarm.ServiceSpec) (
		[]*entity.Setting, error) {
		settings, err := replicaSettings(sources, replica, primary, spec, password, now)
		if err != nil {
			return nil, err
		}
		// Its own directory of the volume, where the primary has its data.
		built, err := uc.volumeService.BuildAppMounts(ctx, db, &volumeservice.BuildAppMountsReq{
			App: replica,
			New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: plan.volumeID,
				Target: spec.DataTarget, VolumeOptions: &volumeservice.AppMountVolumeOptions{}}},
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		svc.TaskTemplate.ContainerSpec.Mounts = built.Mounts
		artifactservice.Attach(svc, plan.artifactMounts)
		return settings, nil
	}
	made, err := uc.appProvisionService.ProvisionApps(ctx, db, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{{
			ProjectID:    app.ProjectID,
			ProjectEnvID: app.ProjectEnvID,
			Key:          key,
			Name:         fmt.Sprintf("%s %s", app.Name, key[len(app.Key)+1:]),
			Status:       base.AppStatusActive,
			// The primary's: out of the app lists and the plan's count, and
			// deleted with it.
			LogicalParentID: app.ID,
			Configure:       configure,
			Deployment: &appprovisionservice.FirstDeployment{
				Source: base.DeploymentTriggerSourceAPI, SourceID: auth.User.ID},
		}},
	})
	return made, hperrors.Wrap(err)
}

// RemoveManagedReplica removes a replica: it goes from the primary's members,
// and its app is deleted, with its directory of the volume.
func (uc *UC) RemoveManagedReplica(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.RemoveManagedReplicaReq,
) (*appmanageddto.RemoveManagedReplicaResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
		if err != nil {
			return err
		}
		pg := loaded.managed.Postgres
		if pg == nil || pg.Member(req.MemberID) == nil {
			return hperrors.NewNotFound("Replica")
		}
		now := timeutil.NowUTC()
		// The base backups taken on it are taken on the database again.
		if pg.BackupFrom.ID == req.MemberID {
			if err = uc.moveBaseBackupsHome(ctx, db, loaded, now); err != nil {
				return err
			}
		}
		pg.Members = slices.DeleteFunc(pg.Members, func(member *entity.ManagedMember) bool {
			return member.App.ID == req.MemberID
		})
		if err = uc.saveManagedSetting(ctx, db, loaded.setting, loaded.managed, now); err != nil {
			return err
		}
		if err = uc.recordReplica(ctx, db, auth, loaded.app, "replicaRemoved", req.MemberID); err != nil {
			return err
		}
		// Last: a deletion reaches past the database.
		return uc.deleteVerifier(ctx, db, loaded.app, req.MemberID)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &appmanageddto.RemoveManagedReplicaResp{}, nil
}

func (uc *UC) recordReplica(ctx context.Context, db database.IDB, auth *basedto.Auth, app *entity.App,
	what, replicaID string) error {
	return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIUpdate,
		Section:  auditSectionManaged,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   auditdetail.New().Set(what, replicaID).String(),
	}))
}

// mountPaths are where the app's setting mounts put files.
func (uc *UC) mountPaths(ctx context.Context, db database.IDB, appID string) ([]string, error) {
	mounts, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return mountPathsOf(mounts), nil
}

func mountPathsOf(settings []*entity.Setting) []string {
	var out []string
	for _, setting := range settings {
		if setting.Type != base.SettingTypeAppSettingMount {
			continue
		}
		if data, err := setting.AsAppSettingMount(); err == nil {
			for _, file := range data.Files {
				out = append(out, file.Path)
			}
		}
	}
	return out
}

// replicaKey is the first <primary key>-r<n> no member has.
func replicaKey(primaryKey string, pg *entity.ManagedPostgres) string {
	for n := 1; ; n++ {
		key := projecthelper.CalcAppKey(fmt.Sprintf("%s-r%d", primaryKey, n))
		taken := false
		for _, member := range pg.Members {
			taken = taken || member.Key == key
		}
		if !taken {
			return key
		}
	}
}

func newReplicationPassword() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, replicationPasswordLen)
	random := make([]byte, replicationPasswordLen)
	_, _ = rand.Read(random) // crypto/rand.Read never fails
	for i, b := range random {
		out[i] = letters[int(b)%len(letters)]
	}
	return string(out)
}

// replicaSettings are the primary's settings, the replica's own as the
// verifier's are, keeping the server's command; the replication password as a
// secret file; the standby's variables; and an app-managed setting of its own
// that names its primary and holds its locks, the primary's for what it copied.
func replicaSettings(
	sources []*entity.Setting,
	replica *entity.App,
	primary *loadedManaged,
	spec *managedmodel.ReplicaSpec,
	password string,
	now time.Time,
) ([]*entity.Setting, error) {
	versions := map[base.SettingType]int{
		base.SettingTypeEnvVar: entity.CurrentEnvVarsVersion, base.SettingTypeSecret: entity.CurrentSecretVersion,
		base.SettingTypeAppSettingMount: entity.CurrentAppSettingMountVersion,
		base.SettingTypeAppManaged:      entity.CurrentAppManagedVersion,
	}
	newSetting := func(typ base.SettingType, name string, data entity.SettingData) *entity.Setting {
		setting := &entity.Setting{ID: gofn.Must(ulid.NewStringULID()), Scope: base.ObjectScopeApp,
			ObjectID: replica.ID, Type: typ, Name: name, Status: base.SettingStatusActive, Version: versions[typ],
			CreatedAt: now, UpdatedAt: now}
		setting.MustSetData(data)
		return setting
	}
	ids := make(map[string]string, len(sources))
	out := make([]*entity.Setting, 0, len(sources)+3) //nolint:mnd // the secret, its mount and the member's
	var env *entity.Setting
	for _, source := range sources {
		if !gofn.Contain(replicaSourceTypes, source.Type) {
			continue
		}
		setting := &entity.Setting{ID: gofn.Must(ulid.NewStringULID()), Scope: base.ObjectScopeApp,
			ObjectID: replica.ID, Type: source.Type, Kind: source.Kind, Name: source.Name,
			Status: base.SettingStatusActive, Inheritable: source.Inheritable, Version: source.Version,
			Data: source.Data, CreatedAt: now, UpdatedAt: now}
		ids[source.ID] = setting.ID
		if setting.Type == base.SettingTypeEnvVar {
			env = setting
		}
		out = append(out, setting)
	}
	for _, setting := range out {
		if err := adaptForReplica(setting, ids); err != nil {
			return nil, err
		}
	}
	if env == nil {
		env = newSetting(base.SettingTypeEnvVar, "", &entity.EnvVars{})
		env.Inheritable = true
		out = append(out, env)
	}
	vars, err := env.AsEnvVars()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, pair := range spec.Env {
		vars.Data = append(vars.Data, &entity.EnvVar{Key: pair[0], Value: pair[1], IsLiteral: true})
	}
	if err = env.SetData(vars); err != nil {
		return nil, hperrors.Wrap(err)
	}

	secret := newSetting(base.SettingTypeSecret, spec.SecretName,
		&entity.Secret{Key: spec.SecretName, Value: entity.NewEncryptedField(password)})
	mode, err := fileutil.ParseFileMode(replicaSecretMode)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	secretMount := newSetting(base.SettingTypeAppSettingMount, replicaSecretMountName,
		&entity.AppSettingMount{Source: entity.ObjectID{ID: secret.ID},
			Files: []*entity.AppSettingMountFile{{Part: "value", Path: spec.PasswordFile,
				UID: postgresFileOwner, GID: postgresFileOwner, Mode: mode}}})
	out = append(out, secret, secretMount)

	locks := primary.managed.Locks
	member := &entity.AppManagedSettings{
		Engine:    primary.managed.Engine,
		EngineAPI: primary.managed.EngineAPI,
		Locks: entity.ManagedLocks{Image: locks.Image, ImageLine: locks.ImageLine, Command: locks.Command,
			Mounts: []string{spec.DataTarget}, EnvKeys: slices.Clone(locks.EnvKeys)},
		MemberOf: &entity.ManagedMemberOf{App: entity.ObjectID{ID: primary.app.ID}, Role: entity.ManagedRoleReplica},
	}
	for _, pair := range spec.Env {
		member.Locks.EnvKeys = append(member.Locks.EnvKeys, pair[0])
	}
	for _, id := range locks.Settings {
		if copied, found := ids[id]; found {
			member.Locks.Settings = append(member.Locks.Settings, copied)
		}
	}
	member.Locks.Settings = append(member.Locks.Settings, secret.ID, secretMount.ID)
	out = append(out, newSetting(base.SettingTypeAppManaged, "", member))
	return out, nil
}

func adaptForReplica(setting *entity.Setting, ids map[string]string) error {
	switch setting.Type { //nolint:exhaustive // two of replicaSourceTypes change
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
		// The server's command stays: a standby does not archive. What runs
		// around a deployment is the primary's to run.
		deployment.PreDeploymentCommand, deployment.PostDeploymentCommand = "", ""
		deployment.RestorePointBeforeDeploy = entity.ObjectID{}
		return hperrors.Wrap(setting.SetData(deployment))
	}
	return nil
}

// moveBaseBackupsHome has the engine take the base backups on the database
// again, and stores the nightly job it moved. The engine's setting is changed
// in place, for the caller to store.
func (uc *UC) moveBaseBackupsHome(ctx context.Context, db database.Tx, loaded *loadedManaged, now time.Time) error {
	jobs := make(map[string]*entity.Setting, len(loaded.jobs))
	for _, job := range loaded.jobs {
		jobs[job.ID] = job
	}
	resp, err := loaded.engine.Reconfigure(&managedmodel.ReconfigureReq{App: loaded.app, ManagedID: loaded.setting.ID,
		Managed: loaded.managed, Jobs: jobs, Change: &managedmodel.SettingsChange{BackupFrom: new("")}, Now: now})
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range resp.Changed {
		setting.UpdatedAt = now
		if err = uc.settingRepo.Update(ctx, db, setting,
			bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, setting, false); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}
