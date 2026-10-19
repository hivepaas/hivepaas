package appmanageduc

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// verifierPlan is a verifier app resolved before the transaction: the volume it
// restores into, the key it reads with, and its artifacts, already on its node.
type verifierPlan struct {
	volumeID       string
	nodeID         string
	keyAuth        entity.ObjectID
	secretValues   map[string]string
	spec           managedmodel.VerifierSpec
	artifactMounts []mount.Mount
}

// planVerifier resolves a move of the restore test into a verifier app, or
// returns nil for a request that asks none. It refuses what would refuse it
// later, before anything is downloaded: a plan without the feature, an engine
// without verifier apps, a verifier already there, an archive on a volume, a
// volume not pinned to a node.
func (uc *UC) planVerifier(ctx context.Context, req *appmanageddto.UpdateManagedSettingsReq) (*verifierPlan, error) {
	if req.RestoreTest == nil || req.RestoreTest.Mode != entity.ManagedRestoreTestVerifier {
		return nil, nil
	}
	err := uc.licenseService.CheckFeatures(ctx, []string{string(base.LicenseFeatureManagedDatabases)})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	if err = loaded.requireVerifierApp(); err != nil {
		return nil, err
	}
	state := loaded.managed.Database
	if state == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	if state.RestoreTest.Mode == entity.ManagedRestoreTestVerifier {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedVerifierExists)
	}
	if state.Storage.Placement != entity.ManagedPlacementOffSite {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedVerifierNeedsObjectStorage)
	}

	plan := &verifierPlan{spec: loaded.engine.Verifier()}
	scope := appScope(loaded.app)
	pin, err := uc.memberVolumePin(ctx, scope, req.RestoreTest.Volume.ID, plan.spec.ScratchDir, false)
	if err != nil {
		return nil, err
	}
	plan.volumeID, plan.nodeID = req.RestoreTest.Volume.ID, pin.NodeID
	if keyAuth := req.RestoreTest.KeyAuth; keyAuth != nil && keyAuth.ID != "" {
		storage := state.Storage
		storage.KeyAuth = entity.ObjectID{ID: keyAuth.ID}
		s3, resolveErr := uc.managedService.ResolveS3(ctx, uc.db, scope, storage)
		if resolveErr != nil {
			return nil, hperrors.Wrap(resolveErr)
		}
		plan.keyAuth = storage.KeyAuth
		plan.secretValues = loaded.engine.ArchiveSecrets(&managedmodel.Archive{S3: s3})
	}

	artifacts, err := uc.artifactService.ArtifactsOf(ctx, uc.db, loaded.app.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	plan.artifactMounts, err = uc.artifactService.MountsFor(ctx, artifacts, placementservice.PinnedSpec(pin))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return plan, nil
}

// memberVolumePin is the node a verifier's or a replica's volume is on, which
// is the node it runs on: none for a volume every node reaches, on which its
// artifacts are then had. A volume that keeps a replica's data at target is
// refused when it keeps it in memory; a verifier's scratch may.
func (uc *UC) memberVolumePin(ctx context.Context, scope *entity.ObjectScope, volumeID, target string,
	keepsData bool) (placementservice.VolumePin, error) {
	setting, err := uc.settingRepo.GetByID(ctx, uc.db, scope, base.SettingTypeClusterVolume, volumeID, true)
	if err != nil {
		return placementservice.VolumePin{}, hperrors.Wrap(err)
	}
	volume, err := setting.AsClusterVolume()
	if err != nil {
		return placementservice.VolumePin{}, hperrors.Wrap(err)
	}
	if keepsData && volumeservice.DriverOptsInMemory(volume.DriverOpts) {
		return placementservice.VolumePin{}, hperrors.Wrap(hperrors.ErrAppManagedVolumeInMemory).
			WithParam("Target", target).ExposeParams("Target")
	}
	return placementservice.VolumePin{VolumeName: volumeID, NodeID: volume.NodeID, NodeLabel: volume.NodeLabel}, nil
}

func appScope(app *entity.App) *entity.ObjectScope {
	_, envKey := projecthelper.ParseProjectEnvID(app.ProjectEnvID)
	return entity.NewObjectScopeProjectEnv(app.ProjectID, envKey)
}

// provisionVerifier makes the verifier app, a logical child of the database:
// its settings copied, its volume and artifacts mounted, at no replicas. It
// returns what it made even when it fails, for the caller to clean up.
func (uc *UC) provisionVerifier(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	loaded *loadedManaged,
	plan *verifierPlan,
) (*appprovisionservice.ProvisionAppsResp, error) {
	app := loaded.app
	sources, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectWhereIn("setting.type IN (?)", verifierSourceTypes...),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	now := timeutil.NowUTC()
	configure := func(ctx context.Context, db database.IDB, verifier *entity.App, spec *swarm.ServiceSpec) (
		[]*entity.Setting, error) {
		settings, err := verifierSettings(sources, verifier, now, plan.secretValues, plan.spec)
		if err != nil {
			return nil, err
		}
		// Its own directory of the volume: VolumeOptions are what give it one.
		built, err := uc.volumeService.BuildAppMounts(ctx, db, &volumeservice.BuildAppMountsReq{
			App: verifier,
			New: []*volumeservice.AppMountReq{{Type: mount.TypeVolume, Source: plan.volumeID,
				Target: plan.spec.ScratchDir, VolumeOptions: &volumeservice.AppMountVolumeOptions{}}},
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		spec.TaskTemplate.ContainerSpec.Mounts = built.Mounts
		artifactservice.Attach(spec, plan.artifactMounts)
		// Up only for its test: the job scales it.
		spec.Mode.Replicated.Replicas = new(uint64(0))
		return settings, nil
	}
	made, err := uc.appProvisionService.ProvisionApps(ctx, db, &appprovisionservice.ProvisionAppsReq{
		Apps: []*appprovisionservice.ProvisionAppReq{{
			ProjectID:    app.ProjectID,
			ProjectEnvID: app.ProjectEnvID,
			Key:          projecthelper.CalcAppKey(app.Key + "-verifier"),
			Name:         app.Name + " verifier",
			Status:       base.AppStatusActive,
			// The database's: out of the app lists and the plan's count, and
			// deleted with it.
			LogicalParentID: app.ID,
			Configure:       configure,
			Deployment: &appprovisionservice.FirstDeployment{
				Source: base.DeploymentTriggerSourceAPI, SourceID: auth.User.ID},
		}},
	})
	return made, hperrors.Wrap(err)
}

// deleteVerifier deletes the database's verifier app, with its volume's
// directory: what it restored there is a copy of the database. One already gone
// is nothing to delete.
func (uc *UC) deleteVerifier(ctx context.Context, db database.IDB, app *entity.App, verifierID string) error {
	if verifierID == "" {
		return nil
	}
	verifier, err := uc.appService.LoadApp(ctx, db, app.ProjectID, verifierID, false, false,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
	)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.appService.DeleteApp(ctx, db, verifier, true, true))
}
