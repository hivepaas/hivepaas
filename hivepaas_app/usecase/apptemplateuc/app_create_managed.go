package apptemplateuc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"path"
	"strings"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// Where an archive is relative to the database it keeps.
const (
	managedPlacementOffSite  = entity.ManagedPlacementOffSite
	managedPlacementShared   = entity.ManagedPlacementShared
	managedPlacementSameNode = entity.ManagedPlacementSameNode

	// walgDir is the directory every archive goes in, under its storage's own
	// prefix, so it never mixes with what a repository keeps. An archive is
	// <app key>/<app ID> below it: the key is shared by apps of other
	// environments and projects, and by an app made again by a name whose archive
	// is kept - HivePaaS deletes none.
	walgDir = "walg"
	// encryptionKeyBytes is the libsodium key's size.
	encryptionKeyBytes = 32
)

// managedPlan is what a managed template's app is created with, resolved before
// the transaction: the engine, the document prepared for it, the archive as the
// engine's setting records it, and the artifacts, already on the node.
type managedPlan struct {
	engine         managedmodel.Engine
	prepared       *managedmodel.PrepareReq
	storage        entity.ManagedStorage
	placement      string
	artifacts      *entity.AppArtifactsSettings
	artifactMounts []mount.Mount
}

// planManaged resolves a managed template's request, or returns nil for an
// ordinary template. It prepares the app's rendered document in place, and has
// the artifacts on the data volume's node.
func (uc *UC) planManaged(
	ctx context.Context,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
	restore *managedRestore,
) (*managedPlan, error) {
	main := mainApp(apps)
	for _, target := range apps {
		if target != main && target.rendered != nil && target.rendered.Template.IsManaged() {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("a managed template cannot be a dependency yet")
		}
	}
	if main == nil || !main.rendered.Template.IsManaged() {
		if req.Managed != nil || restore != nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
				WithExtraDetail("managed: the template is not managed")
		}
		return nil, nil
	}
	tmpl := main.rendered.Template
	if req.Managed == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
			WithExtraDetail("managed: where the database archives is required")
	}
	engine, found := uc.managedService.Engine(tmpl.Managed.Engine)
	if minAPI, maxAPI := engineRange(engine, found); tmpl.Managed.EngineAPI < minAPI || tmpl.Managed.EngineAPI > maxAPI {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).WithExtraDetail(
			"engine %s API %d is not one this HivePaaS has", tmpl.Managed.Engine, tmpl.Managed.EngineAPI)
	}

	scope := uc.envScope(req)
	pin, err := uc.dataVolumePin(ctx, scope, main)
	if err != nil {
		return nil, err
	}
	planned := &managedPlan{engine: engine}
	archive, err := uc.resolveArchive(ctx, scope, req.Managed.Storage, pin, path.Join(main.key(), main.id), planned)
	if err != nil {
		return nil, err
	}
	options := managedOptions(tmpl.Managed.Defaults(), req.Managed)
	if err = checkSchedules(options); err != nil {
		return nil, err
	}
	if options.DumpRepo, err = uc.dumpRepo(ctx, scope, req.Managed.Dump, planned.storage); err != nil {
		return nil, err
	}
	planned.prepared = &managedmodel.PrepareReq{
		Doc:           main.result.Doc,
		Version:       main.result.Version.Name,
		Archive:       archive,
		EncryptionKey: newEncryptionKey(),
		Options:       options,
	}
	if restore != nil {
		planned.prepared.Restore = restore.source
	}
	if err = engine.PrepareDoc(planned.prepared); err != nil {
		return nil, hperrors.Wrap(err)
	}

	planned.artifacts = artifactservice.FromTemplate(tmpl.Artifacts)
	planned.artifactMounts, err = uc.artifactService.MountsFor(ctx, planned.artifacts, placementservice.PinnedSpec(pin))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return planned, nil
}

func engineRange(engine managedmodel.Engine, found bool) (int, int) {
	if !found {
		return 1, 0
	}
	return engine.APIRange()
}

// mainApp is the app the request is about: the one no other app owns.
func mainApp(apps []*appToProvision) *appToProvision {
	for _, target := range apps {
		if target.logicalParentID == "" {
			return target
		}
	}
	return nil
}

func (uc *UC) envScope(req *apptemplatedto.CreateAppFromTemplateReq) *entity.ObjectScope {
	_, envKey := projecthelper.ParseProjectEnvID(req.ProjectEnvID)
	return entity.NewObjectScopeProjectEnv(req.ProjectID, envKey)
}

// dataVolumePin is the node the app's data is on: the pin its volumes share,
// none when every node reaches them - shared storage, by the volume's own claim.
// A database keeps its data on a volume, never in memory.
func (uc *UC) dataVolumePin(
	ctx context.Context,
	scope *entity.ObjectScope,
	main *appToProvision,
) (placementservice.VolumePin, error) {
	var pins []placementservice.VolumePin
	if storage := main.result.Doc.Deployment.Storage; storage != nil {
		for target, m := range storage.Mounts {
			if m.Type != mount.TypeVolume {
				continue
			}
			volume, err := uc.clusterVolume(ctx, scope, m.Source)
			if err != nil {
				return placementservice.VolumePin{}, err
			}
			if volumeservice.DriverOptsInMemory(volume.DriverOpts) {
				return placementservice.VolumePin{}, hperrors.Wrap(hperrors.ErrAppManagedVolumeInMemory).
					WithParam("Target", target).ExposeParams("Target")
			}
			pins = append(pins, placementservice.VolumePin{VolumeName: m.Source,
				NodeID: volume.NodeID, NodeLabel: volume.NodeLabel})
		}
	}
	if len(pins) == 0 {
		return placementservice.VolumePin{}, hperrors.Wrap(hperrors.ErrAppManagedVolumeNotPinned).
			WithExtraDetail("it mounts no volume")
	}
	if _, conflict := placementservice.VolumePinConstraint(pins); conflict != nil {
		return placementservice.VolumePin{}, hperrors.Wrap(hperrors.ErrAppManagedVolumeNotPinned).
			WithExtraDetail("its volumes are pinned to different nodes")
	}
	for _, pin := range pins {
		if pin.NodeID != "" || pin.NodeLabel != "" {
			return pin, nil
		}
	}
	return placementservice.VolumePin{}, nil
}

func (uc *UC) clusterVolume(ctx context.Context, scope *entity.ObjectScope, id string) (*entity.ClusterVolume, error) {
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, scope, base.SettingTypeClusterVolume, id, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	volume, err := setting.AsClusterVolume()
	return volume, hperrors.Wrap(err)
}

// resolveArchive is where the engine archives: the credentials decrypted for the
// engine, the references for its setting, and where it is from the database.
// planned is given the latter two.
func (uc *UC) resolveArchive(
	ctx context.Context,
	scope *entity.ObjectScope,
	req apptemplatedto.ManagedStorageReq,
	dataPin placementservice.VolumePin,
	name string,
	planned *managedPlan,
) (*managedmodel.Archive, error) {
	prefix := path.Join(walgDir, name)
	if req.BackupRepo != nil && req.BackupRepo.ID != "" {
		setting, err := uc.settingRepo.GetByID(ctx, uc.db, scope, base.SettingTypeBackupRepo, req.BackupRepo.ID, true)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		repo, err := setting.AsBackupRepo()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		planned.storage.BackupRepo = entity.ObjectID{ID: setting.ID}
		prefix = path.Join(repo.StoragePrefix, walgDir, name)
		if repo.Volume.ID != "" {
			return uc.volumeArchive(ctx, scope, repo.Volume.ID, prefix, dataPin, planned)
		}
	} else if req.CloudStorage != nil {
		planned.storage.CloudStorage = entity.ObjectID{ID: req.CloudStorage.ID}
	}
	planned.storage.Prefix = prefix
	planned.placement = managedPlacementOffSite
	planned.storage.Placement = planned.placement
	if req.KeyAuth != nil && req.KeyAuth.ID != "" {
		planned.storage.KeyAuth = entity.ObjectID{ID: req.KeyAuth.ID}
	}
	s3, err := uc.managedService.ResolveS3(ctx, uc.db, scope, planned.storage)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &managedmodel.Archive{S3: s3}, nil
}

// volumeArchive archives into a volume repository's directory, which the
// database's node has to be able to mount: the same node, or shared storage. A
// database on shared storage has no node of its own, and archives off-site.
func (uc *UC) volumeArchive(
	ctx context.Context,
	scope *entity.ObjectScope,
	volumeID, prefix string,
	dataPin placementservice.VolumePin,
	planned *managedPlan,
) (*managedmodel.Archive, error) {
	if dataPin.NodeID == "" && dataPin.NodeLabel == "" {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedSharedVolumeNeedsObjectStorage)
	}
	volume, err := uc.clusterVolume(ctx, scope, volumeID)
	if err != nil {
		return nil, err
	}
	switch {
	case volume.NodeID == "" && volume.NodeLabel == "":
		planned.placement = managedPlacementShared
	case (volume.NodeID != "" && volume.NodeID == dataPin.NodeID) ||
		(volume.NodeLabel != "" && volume.NodeLabel == dataPin.NodeLabel):
		planned.placement = managedPlacementSameNode
	default:
		return nil, hperrors.Wrap(hperrors.ErrAppManagedArchiveUnreachable)
	}
	planned.storage.Prefix = prefix
	planned.storage.Placement = planned.placement
	return &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{VolumeID: volumeID, Subpath: prefix}}, nil
}

// managedOptions are the request's options over the template's defaults; what
// both leave out is the engine's.
func managedOptions(
	defaults *templatemodel.ManagedDefaults,
	req *apptemplatedto.ManagedCreateReq,
) *managedmodel.Options {
	options := &managedmodel.Options{}
	if defaults != nil {
		options.ArchiveTimeout = defaults.ArchiveTimeout.Duration
		options.RetainFull = defaults.RetainFull
		options.BaseBackupSchedule = defaults.BaseBackupSchedule
		options.RestoreTestSchedule = defaults.RestoreTestSchedule
	}
	options.ArchiveTimeout = gofn.Coalesce(time.Duration(req.ArchiveTimeout), options.ArchiveTimeout)
	options.RetainFull = gofn.Coalesce(req.RetainFull, options.RetainFull)
	options.BaseBackupSchedule = gofn.Coalesce(strings.TrimSpace(req.BaseBackupSchedule), options.BaseBackupSchedule)
	options.RestoreTestSchedule = gofn.Coalesce(strings.TrimSpace(req.RestoreTestSchedule),
		options.RestoreTestSchedule)
	options.RestoreTestQuery = strings.TrimSpace(req.RestoreTestQuery)
	if req.Dump != nil {
		options.DumpSchedule = strings.TrimSpace(req.Dump.Schedule)
	}
	return options
}

// dumpRepo is the backup repository the database is dumped into: the one asked
// for, none when null is asked for, and when dumps are not mentioned, the
// archive's own repository - none for an archive in cloud storage.
func (uc *UC) dumpRepo(
	ctx context.Context,
	scope *entity.ObjectScope,
	req *apptemplatedto.ManagedDumpReq,
	storage entity.ManagedStorage,
) (string, error) {
	if req == nil {
		return storage.BackupRepo.ID, nil
	}
	if req.Repo == nil || req.Repo.ID == "" {
		return "", nil
	}
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, scope, base.SettingTypeBackupRepo, req.Repo.ID, true)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if !setting.IsActive() {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
			WithExtraDetail("managed.dump.repo: the backup repository is not active")
	}
	return setting.ID, nil
}

// checkSchedules refuses a schedule the jobs could not run by, before anything
// is created with it.
func checkSchedules(options *managedmodel.Options) error {
	for field, cron := range map[string]string{
		"baseBackupSchedule": options.BaseBackupSchedule, "restoreTestSchedule": options.RestoreTestSchedule,
		"dump.schedule": options.DumpSchedule,
	} {
		if cron == "" {
			continue
		}
		if _, err := (&entity.SchedJobSchedule{CronExpr: cron}).ParseCronExpr(); err != nil {
			return hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
				WithExtraDetail("managed.%s %q is not a cron expression", field, cron)
		}
	}
	return nil
}

func newEncryptionKey() string {
	key := make([]byte, encryptionKeyBytes)
	_, _ = rand.Read(key) // crypto/rand.Read never fails
	return hex.EncodeToString(key)
}

// attach gives the plan to the app the request is about. A nil plan attaches
// nothing.
func (p *managedPlan) attach(apps []*appToProvision) {
	if p == nil {
		return
	}
	if main := mainApp(apps); main != nil {
		main.managed = p
	}
}

// configure runs inside the transaction, on the settings the document was built
// into: the engine finishes them and adds its own, and the app is given its
// artifacts' setting and mounts.
func (p *managedPlan) configure(
	app *entity.App,
	spec *swarm.ServiceSpec,
	settings []*entity.Setting,
	timeNow time.Time,
) ([]*entity.Setting, error) {
	configured, err := p.engine.Configure(&managedmodel.ConfigureReq{
		App:      app,
		Prepared: p.prepared,
		Storage:  p.storage,
		Settings: settings,
		Now:      timeNow,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := configured.Settings
	if p.artifacts != nil {
		artifacts := &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeApp,
			ObjectID:  app.ID,
			Type:      base.SettingTypeAppArtifacts,
			Status:    base.SettingStatusActive,
			Version:   entity.CurrentAppArtifactsVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		artifacts.MustSetData(p.artifacts)
		out = append(out, artifacts)
		artifactservice.Attach(spec, p.artifactMounts)
	}
	return out, nil
}
