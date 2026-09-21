package appsettingsuc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppStorageSettingsReq,
) (*appsettingsdto.GetAppStorageSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// What each mount turned out to reach, and the apps of the environment to
	// name the owner of a directory that is not this app's. Neither is fatal:
	// without them every mount reads as the app's own, which is what the screen
	// showed before an app could be given another app's storage.
	mounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	descs, err := uc.volumeService.DescribeAppMounts(ctx, uc.db, app, mounts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	appsByKey, err := uc.appsInEnvByKey(ctx, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appsettingsdto.StorageSettingsTransformInput{
		App:                app,
		Service:            service,
		MountKeyCalculator: uc.calcMountKey,
		MountDescs:         descs,
		AppsByKey:          appsByKey,
	}

	resp, err := appsettingsdto.TransformStorageSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BorrowedBy = uc.findStorageBorrowers(ctx, app, appsByKey)

	return &appsettingsdto.GetAppStorageSettingsResp{
		Data: resp,
	}, nil
}

// findStorageBorrowers is the apps of this environment that mount a directory of
// this one's.
//
// It reads their services, because a mount lives in the service spec and nowhere
// else. An app whose service cannot be read is left out rather than failing the
// screen: an incomplete list is worth more than an error page, and the mounts of
// the app being read - the part this screen is actually for - are already in
// hand by the time this runs.
func (uc *UC) findStorageBorrowers(
	ctx context.Context,
	app *entity.App,
	appsByKey map[string]*entity.App,
) []*appsettingsdto.MountBorrower {
	var borrowers []*appsettingsdto.MountBorrower

	for _, other := range appsByKey {
		if other.ID == app.ID || other.ServiceID == "" {
			continue
		}
		// Every app here is in the same project and environment as the app being
		// read, which is what its directory prefix is built from.
		other.Project, other.ProjectEnv = app.Project, app.ProjectEnv

		service, err := uc.clusterService.ServiceInspect(ctx, other.ServiceID, false)
		if err != nil || service == nil || service.Spec.TaskTemplate.ContainerSpec == nil {
			continue
		}
		mounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
		descs, err := uc.volumeService.DescribeAppMounts(ctx, uc.db, other, mounts)
		if err != nil {
			continue
		}

		for i, desc := range descs {
			if desc == nil || desc.Own || desc.AppKey != app.Key {
				continue
			}
			borrowers = append(borrowers, &appsettingsdto.MountBorrower{
				AppID:   other.ID,
				Name:    other.Name,
				Target:  mounts[i].Target,
				Subpath: desc.Subpath,
				Write:   !mounts[i].ReadOnly,
			})
		}
	}

	sort.Slice(borrowers, func(i, j int) bool {
		if borrowers[i].Name != borrowers[j].Name {
			return borrowers[i].Name < borrowers[j].Name
		}
		return borrowers[i].Target < borrowers[j].Target
	})
	return borrowers
}

// appsInEnvByKey is every app beside this one, by the key that appears in a
// directory path inside a volume.
func (uc *UC) appsInEnvByKey(ctx context.Context, app *entity.App) (map[string]*entity.App, error) {
	apps, _, err := uc.appRepo.List(ctx, uc.db, app.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", app.ProjectEnvID),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	byKey := make(map[string]*entity.App, len(apps))
	for _, item := range apps {
		byKey[item.Key] = item
	}
	return byKey, nil
}

func (uc *UC) calcMountKey(mnt *mount.Mount) string {
	key := fmt.Sprintf("type:%v:src:%v:target:%v", mnt.Type, mnt.Source, mnt.Target)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}
