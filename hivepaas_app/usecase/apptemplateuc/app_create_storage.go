package apptemplateuc

import (
	"context"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// storageFinding is one app of this request whose directory in a volume already
// holds something.
//
// A template that creates eight apps produces up to eight of these, which is why
// they are a list to be read rather than a refusal: the answer people want is
// one decision about all of them, not eight.
type storageFinding struct {
	AppName    string
	AppKey     string
	IsDatabase bool
	VolumeID   string
	VolumeName string
	Path       string
}

// storagePlan is what the apps of a request would be given, and what is already
// there. The queries are kept because clearing the directories asks for exactly
// the same thing the inspection did.
type storagePlan struct {
	Request  *volumeservice.InspectAppStorageReq
	Findings []*storageFinding
	// Unchecked is what nothing could be seen of: storage on a node that could
	// not be reached, or a volume with nothing on a filesystem to look at. It is
	// reported rather than dropped, because an empty answer that means "nothing
	// was seen" reads exactly like one that means "there is nothing there".
	Unchecked []*storageFinding
}

// planStorage asks, for every app this request would create, whether the
// directory it would be given already holds something.
//
// It reads the mounts the template rendered rather than re-deriving anything:
// the key of each app is the one the creation itself will use, and a mount names
// the volume setting it will be made from.
func (uc *UC) planStorage(
	ctx context.Context,
	db database.IDB,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) (*storagePlan, error) {
	_, envKey := projecthelper.ParseProjectEnvID(req.ProjectEnvID)
	if envKey == "" {
		return &storagePlan{}, nil
	}
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	inspectReq := &volumeservice.InspectAppStorageReq{
		Scope: entity.NewObjectScopeProjectEnv(req.ProjectID, envKey),
	}
	byKey := map[string]*appToProvision{}
	for _, app := range apps {
		key := app.key()
		byKey[key] = app
		inspectReq.Queries = append(inspectReq.Queries, appStorageQueries(app, key, project, envKey)...)
	}
	if len(inspectReq.Queries) == 0 {
		return &storagePlan{}, nil
	}

	resp, err := uc.volumeService.InspectAppStorage(ctx, db, inspectReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	plan := &storagePlan{Request: inspectReq}
	for _, state := range resp.States {
		app := byKey[state.AppKey]
		if app == nil || (state.Checked && !state.HasData()) {
			continue
		}
		finding := &storageFinding{
			AppName:    app.name,
			AppKey:     state.AppKey,
			IsDatabase: appIsDatabase(app),
			VolumeID:   state.VolumeID,
			VolumeName: state.VolumeName,
			Path:       state.Path,
		}
		if state.Checked {
			plan.Findings = append(plan.Findings, finding)
			continue
		}
		plan.Unchecked = append(plan.Unchecked, finding)
	}
	return plan, nil
}

// appStorageQueries is what one app would be given: its own directory in each
// volume it mounts.
//
// A mount that names another app is left out. It reaches somebody else's
// directory on purpose - a file manager over the database beside it - and what
// is in there is not a leftover of this app.
func appStorageQueries(
	app *appToProvision,
	key string,
	project *entity.Project,
	envKey string,
) []*volumeservice.AppStorageQuery {
	doc := app.result.Doc
	if doc == nil || doc.Deployment == nil || doc.Deployment.Storage == nil {
		return nil
	}

	planned := &entity.App{
		Key:        key,
		Project:    project,
		ProjectEnv: &entity.ProjectEnv{Key: envKey},
	}
	queries := make([]*volumeservice.AppStorageQuery, 0, len(doc.Deployment.Storage.Mounts))
	for _, mnt := range doc.Deployment.Storage.Mounts {
		if mnt.SourceApp != nil || mnt.Source == "" {
			continue
		}
		if mnt.Type != mount.TypeVolume && mnt.Type != mount.TypeCluster {
			continue
		}
		queries = append(queries, &volumeservice.AppStorageQuery{
			AppKey:   key,
			App:      planned,
			VolumeID: mnt.Source,
			Subpath:  mountOwnSubpath(&mnt),
		})
	}
	return queries
}

func mountOwnSubpath(mnt *specmodel.Mount) string {
	switch {
	case mnt.VolumeOptions != nil:
		return mnt.VolumeOptions.Subpath
	case mnt.ClusterOptions != nil:
		return mnt.ClusterOptions.Subpath
	default:
		return ""
	}
}

// appIsDatabase reads the category off the kind block the template rendered. A
// database is the app a leftover directory is certain to break: the password it
// is created with is generated afresh, while the cluster on disk keeps the one
// it was initialized with.
func appIsDatabase(app *appToProvision) bool {
	doc := app.result.Doc
	if doc == nil {
		return false
	}
	kind, _ := doc.Settings["kind"].(map[string]any)
	if kind == nil {
		return false
	}
	category, _ := kind["category"].(string)
	return base.AppCategory(category) == base.AppCategoryDatabase
}
