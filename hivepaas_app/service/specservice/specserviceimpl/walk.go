package specserviceimpl

import (
	"context"
	"sort"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// scopeUnit is one scope's worth of exportable settings, gathered before any
// document is written.
//
// Collecting everything first is what makes references resolve regardless of
// walk order. Assembling as the walk goes would leave an env setting unable to
// reference a project setting simply because the project document had not been
// reached yet - a correctness bug that would look like a nondeterministic one.
type scopeUnit struct {
	path     string
	settings []*entity.Setting
}

// buildBundle gathers every scope, indexes them all, and only then writes the
// documents.
func (s *service) buildBundle(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
) (*specmodel.Bundle, error) {
	bundle := &specmodel.Bundle{Files: map[string][]byte{}, Report: &specmodel.Report{}}

	netNames, err := s.loadNetworkNames(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	tree, err := s.gather(ctx, db, req.Scope, bundle.Report)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Pass one: every setting in the export gets a path before anything is
	// written, so a reference from any scope to any other resolves.
	index := newRefIndex()
	for _, unit := range tree.units {
		indexSettings(index, unit.settings, unit.path)
	}

	// Pass one and a half: a reference to a setting the export does not hold - a
	// project exported alone whose app uses a global certificate - is written as
	// an external reference, which says what finds its target again elsewhere.
	if err = s.indexExternalRefs(ctx, db, tree, index); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Pass two: write the documents.
	if err = s.writeDocs(ctx, db, req, tree, netNames, index, bundle); err != nil {
		return nil, hperrors.Wrap(err)
	}

	bundle.Manifest = &specmodel.Manifest{
		APIVersion:        specmodel.APIVersion,
		Kind:              specmodel.KindSpec,
		ExportedAt:        timeutil.NowUTC(),
		SourceAppVersion:  base.StableVersion.AppVersion,
		SourceVersionCode: base.CurrentVersion,
		Scope:             specScopeName(req.Scope.ScopeType),
		SecretsMode:       req.SecretsMode,
		Files:             sortedFilenames(bundle.Files),
	}
	return bundle, nil
}

// exportTree is everything the walk found, in the order documents are written.
type exportTree struct {
	global   *scopeUnit
	projects []*projectUnit
	units    []*scopeUnit
}

type projectUnit struct {
	project *entity.Project
	unit    *scopeUnit
	envs    []*envUnit
}

type envUnit struct {
	env  *entity.ProjectEnv
	unit *scopeUnit
	apps []*appUnit
}

type appUnit struct {
	app  *entity.App
	unit *scopeUnit
}

func (s *service) gather(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	report *specmodel.Report,
) (*exportTree, error) {
	tree := &exportTree{}
	add := func(unit *scopeUnit) *scopeUnit {
		tree.units = append(tree.units, unit)
		return unit
	}

	if scope.ScopeType == base.ObjectScopeGlobal {
		settings, err := s.loadOwned(ctx, db,
			[]base.ObjectScopeType{base.ObjectScopeGlobal, base.ObjectScopeHivepaas}, "")
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		tree.global = add(&scopeUnit{
			path: "global", settings: selectSettings(settings, "global", report),
		})
	}

	projects, err := s.projectsInScope(ctx, db, scope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, project := range projects {
		projUnit, err := s.gatherProject(ctx, db, scope, project, add, report)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		tree.projects = append(tree.projects, projUnit)
	}
	return tree, nil
}

func (s *service) gatherProject(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	project *entity.Project,
	add func(*scopeUnit) *scopeUnit,
	report *specmodel.Report,
) (*projectUnit, error) {
	path := "projects/" + project.Key

	settings, err := s.loadOwned(ctx, db,
		[]base.ObjectScopeType{base.ObjectScopeProject}, project.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &projectUnit{
		project: project,
		unit:    add(&scopeUnit{path: path, settings: selectSettings(settings, path, report)}),
	}

	envs, _, err := s.projectEnvRepo.List(ctx, db, project.ID, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	sort.SliceStable(envs, func(i, j int) bool {
		if envs[i].Index != envs[j].Index {
			return envs[i].Index < envs[j].Index
		}
		return envs[i].Key < envs[j].Key
	})

	for _, env := range envs {
		if scope.ProjectEnvID != "" && env.ID != scope.ProjectEnvID {
			continue
		}
		envOut, err := s.gatherEnv(ctx, db, scope, project, env, add, report)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.envs = append(out.envs, envOut)
	}
	return out, nil
}

func (s *service) gatherEnv(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	project *entity.Project,
	env *entity.ProjectEnv,
	add func(*scopeUnit) *scopeUnit,
	report *specmodel.Report,
) (*envUnit, error) {
	path := "projects/" + project.Key + "/envs/" + env.Key

	settings, err := s.loadOwned(ctx, db,
		[]base.ObjectScopeType{base.ObjectScopeProjectEnv}, env.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := &envUnit{
		env:  env,
		unit: add(&scopeUnit{path: path, settings: selectSettings(settings, path, report)}),
	}

	apps, _, err := s.appRepo.List(ctx, db, project.ID, nil,
		bunex.SelectWhere("app.project_env_id = ?", env.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	for _, app := range selectApps(apps, path, report) {
		if scope.AppID != "" && app.ID != scope.AppID {
			continue
		}
		appPath := path + "/apps/" + app.Key

		appSettings, err := s.loadOwned(ctx, db,
			[]base.ObjectScopeType{base.ObjectScopeApp}, app.ID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		out.apps = append(out.apps, &appUnit{
			app: app,
			unit: add(&scopeUnit{
				path: appPath, settings: selectSettings(appSettings, appPath, report),
			}),
		})
	}
	return out, nil
}

func (s *service) writeDocs(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	tree *exportTree,
	netNames map[string]string,
	index *refIndex,
	bundle *specmodel.Bundle,
) error {
	if tree.global != nil {
		assembled, err := s.assembleScope(tree.global.settings, index, req.SecretsMode)
		if err != nil {
			return hperrors.Wrap(err)
		}
		doc := &specmodel.GlobalDoc{
			DocHeader: specmodel.NewDocHeader(specScopeName(base.ObjectScopeGlobal)),
			Settings:  assembled,
		}
		if err = addFile(bundle, globalFilename, doc); err != nil {
			return hperrors.Wrap(err)
		}
	}

	for _, projUnit := range tree.projects {
		envNames := make([]string, 0, len(projUnit.envs))
		for _, envUnit := range projUnit.envs {
			if err := s.writeEnvDoc(ctx, db, req, projUnit, envUnit, netNames, index, bundle); err != nil {
				return hperrors.Wrap(err)
			}
			envNames = append(envNames, envUnit.env.Key+".yaml")
		}

		assembled, err := s.assembleScope(projUnit.unit.settings, index, req.SecretsMode)
		if err != nil {
			return hperrors.Wrap(err)
		}
		doc := &specmodel.ProjectDoc{
			DocHeader: specmodel.NewDocHeader(specScopeName(base.ObjectScopeProject)),
			Project:   projUnit.project.Key,
			ID:        projUnit.project.ID,
			Name:      projUnit.project.Name,
			Note:      projUnit.project.Note,
			Owner:     projectOwner(projUnit.project),
			Envs:      envNames,
			Settings:  assembled,
		}
		if err = addFile(bundle, projUnit.unit.path+"/"+projectFilename, doc); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *service) writeEnvDoc(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	projUnit *projectUnit,
	env *envUnit,
	netNames map[string]string,
	index *refIndex,
	bundle *specmodel.Bundle,
) error {
	appDocs := map[string]*specmodel.AppDoc{}
	for _, app := range env.apps {
		// Where an app's directory is inside a volume depends on its project and
		// environment, which the app row alone does not carry.
		app.app.Project, app.app.ProjectEnv = projUnit.project, env.env
		doc, err := s.buildAppDoc(ctx, db, req, app, netNames, index, bundle.Report)
		if err != nil {
			return hperrors.Wrap(err)
		}
		appDocs[app.app.Key] = doc
	}

	assembled, err := s.assembleScope(env.unit.settings, index, req.SecretsMode)
	if err != nil {
		return hperrors.Wrap(err)
	}

	doc := &specmodel.EnvDoc{
		DocHeader: specmodel.NewDocHeader(specScopeName(base.ObjectScopeProjectEnv)),
		Project:   projUnit.project.Key,
		Env:       env.env.Key,
		Name:      env.env.Name,
		Color:     env.env.Color,
		Index:     env.env.Index,
		Apps:      appDocs,
		Settings:  assembled,
	}
	return addFile(bundle, env.unit.path+".yaml", doc)
}

func (s *service) buildAppDoc(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	app *appUnit,
	netNames map[string]string,
	index *refIndex,
	report *specmodel.Report,
) (*specmodel.AppDoc, error) {
	assembled, err := s.assembleScope(app.unit.settings, index, req.SecretsMode)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	doc := &specmodel.AppDoc{
		App:      app.app.Key,
		ID:       app.app.ID,
		Name:     app.app.Name,
		Status:   string(app.app.Status),
		Note:     app.app.Note,
		Settings: assembled,
	}

	// The app-deployment setting is lifted out of the flat block map and placed
	// beside the Swarm-derived blocks, which is where a reader expects it.
	sourceBlock := specmodel.SingletonBlockName(base.SettingTypeAppDeployment)
	deployment := &specmodel.Deployment{}
	if source, ok := assembled[sourceBlock]; ok {
		if body, isMap := source.(map[string]any); isMap {
			deployment.Source = body
		}
		delete(assembled, sourceBlock)
	}

	if err = s.addSwarmBlocks(ctx, db, app.app, app.unit.path, netNames, index, deployment, report); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if deployment.Source != nil || deployment.Container != nil {
		doc.Deployment = deployment
	}
	return doc, nil
}

// addSwarmBlocks fills the parts of a deployment that live in the Swarm service
// rather than in settings.
//
// An app that has never been deployed has no service to read. That is a real
// state rather than an edge case - two of five user apps in a development
// installation are in it - so the blocks are simply absent and the report says
// why, rather than the export failing.
func (s *service) addSwarmBlocks(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	scopePath string,
	netNames map[string]string,
	index *refIndex,
	deployment *specmodel.Deployment,
	report *specmodel.Report,
) error {
	if app.ServiceID == "" {
		report.Add(specmodel.Issue{
			Severity: specmodel.SeverityFixable,
			Code:     specmodel.CodeServiceUnavailable,
			Path:     scopePath,
			Action:   "the app has never been deployed; its settings are exported alone",
		})
		return nil
	}

	svc, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		// Swallowed on purpose. One unreadable service must not fail an export
		// of forty apps: the operator gets everything that could be read, plus
		// a report entry naming what could not. That is the same leniency the
		// import contract specifies, applied at the other end.
		report.Add(specmodel.Issue{
			Severity: specmodel.SeverityFixable,
			Code:     specmodel.CodeServiceUnavailable,
			Path:     scopePath,
			Detail:   map[string]any{"serviceId": app.ServiceID},
			Action:   "the swarm service could not be read; its settings are exported alone",
		})
		return nil //nolint:nilerr // reported rather than raised; see above
	}

	mapped := mapSwarmService(svc, netNames)
	if mapped == nil {
		return nil
	}
	storage, err := s.mapStorageOf(ctx, db, app, svc, index)
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Container = mapped.Container
	deployment.Resources = mapped.Resources
	deployment.Storage = storage
	deployment.Networks = mapped.Networks
	deployment.Service = mapped.Service
	return nil
}

// mapStorageOf reads an app's mounts back into what built them. volumeservice
// says which volume and whose directory each one reaches - the same answer the
// storage screen shows - and each volume is named the way the bundle can: by
// its path when the export holds it, by an external reference when it does not.
func (s *service) mapStorageOf(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	svc *swarm.Service,
	index *refIndex,
) (*specmodel.Storage, error) {
	task := &svc.Spec.TaskTemplate
	if task.ContainerSpec == nil || len(task.ContainerSpec.Mounts) == 0 {
		return nil, nil
	}
	descs, err := s.volumeService.DescribeAppMounts(ctx, db, app, task.ContainerSpec.Mounts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	type volumeName struct {
		path     string
		external *specmodel.ExternalRef
	}
	names := map[string]volumeName{}
	for _, desc := range descs {
		if desc == nil || desc.VolumeID == "" {
			continue
		}
		if _, done := names[desc.VolumeID]; done {
			continue
		}
		name := volumeName{path: index.paths[desc.VolumeID]}
		if name.path == "" {
			if name.external, err = s.externalRef(ctx, db, index, desc.VolumeID); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
		names[desc.VolumeID] = name
	}
	return mapAppStorage(task, descs, func(volumeID string) (string, *specmodel.ExternalRef) {
		return names[volumeID].path, names[volumeID].external
	})
}

// loadNetworkNames maps Docker network id to name.
//
// The cluster-network settings sync writes carry both - RefID is the Docker id,
// Name is the Docker name - so this needs no round trip to Docker.
func (s *service) loadNetworkNames(
	ctx context.Context,
	db database.IDB,
) (map[string]string, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNetwork),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	names := make(map[string]string, len(settings))
	for _, setting := range settings {
		if setting.RefID != "" && setting.Name != "" {
			names[setting.RefID] = setting.Name
		}
	}
	return names, nil
}

// loadOwnedFromRepo is the production settingLoader.
//
// It passes a nil scope on purpose. The repository's own scope filters widen a
// query to include what an outer scope defines, which is right for reading
// configuration and wrong for exporting it: each scope's document carries only
// what that scope defines.
func (s *service) loadOwnedFromRepo(
	ctx context.Context,
	db database.IDB,
	scopes []base.ObjectScopeType,
	objectID string,
) ([]*entity.Setting, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.scope IN (?)", scopes...),
	}
	if objectID == "" {
		opts = append(opts, bunex.SelectWhere("setting.object_id IS NULL"))
	} else {
		opts = append(opts, bunex.SelectWhere("setting.object_id = ?", objectID))
	}

	settings, _, err := s.settingRepo.List(ctx, db, nil, nil, opts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return settings, nil
}

func (s *service) loadByIDsFromRepo(
	ctx context.Context,
	db database.IDB,
	ids []string,
) ([]*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhereIn("setting.id IN (?)", ids...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return settings, nil
}

// indexExternalRefs registers every setting an exported setting references but
// the export does not hold. replaceExternalRefs then writes each reference to it
// as an external one. A reference to a setting that no longer exists is left as
// the id it is, for import to report - and so is one GetRefObjectIDs does not
// report, such as the certificate of a disabled domain, which holds no
// reservation and so is not counted as a reference.
func (s *service) indexExternalRefs(
	ctx context.Context,
	db database.IDB,
	tree *exportTree,
	index *refIndex,
) error {
	var ids []string
	seen := map[string]bool{}
	for _, unit := range tree.units {
		for _, setting := range unit.settings {
			refs, err := setting.GetRefObjectIDs()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if refs == nil {
				continue
			}
			for _, id := range refs.RefSettingIDs {
				if id == "" || seen[id] || index.paths[id] != "" {
					continue
				}
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	settings, err := s.loadByIDs(ctx, db, ids)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range settings {
		index.addExternal(setting)
	}
	return nil
}

// externalRef is the external reference for one setting the export does not
// hold, loaded the first time it is asked for. It is nil when no such setting
// exists. Volumes need it one at a time: which ones are mounted is learned only
// from the services, after the settings' external references were registered.
func (s *service) externalRef(
	ctx context.Context,
	db database.IDB,
	index *refIndex,
	id string,
) (*specmodel.ExternalRef, error) {
	if ref := index.external[id]; ref != nil {
		return ref, nil
	}
	settings, err := s.loadByIDs(ctx, db, []string{id})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, setting := range settings {
		index.addExternal(setting)
	}
	return index.external[id], nil
}

func (s *service) projectsInScope(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
) ([]*entity.Project, error) {
	switch scope.ScopeType {
	case base.ObjectScopeGlobal:
		projects, _, err := s.projectRepo.List(ctx, db, nil, withProjectOwner())
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return selectProjects(projects), nil

	case base.ObjectScopeProject, base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		project, err := s.projectRepo.GetByID(ctx, db, scope.ProjectID, withProjectOwner())
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return selectProjects([]*entity.Project{project}), nil

	case base.ObjectScopeHivepaas:
		// The hivepaas scope's settings travel in global.yaml; it owns no
		// projects of its own.
		return nil, nil

	case base.ObjectScopeUser:
		// Nothing there but api-key, which is skipped, so the scope is empty by
		// construction.
		return nil, nil

	default:
		return nil, nil
	}
}

// withProjectOwner loads the user who owns each project, for the email a bundle
// carries. It excludes the columns every other reader of a user excludes.
func withProjectOwner() bunex.SelectQueryOption {
	return bunex.SelectRelation("Owner",
		bunex.SelectExcludeColumns(entity.UserDefaultExcludeColumns...),
	)
}

// projectOwner is a project's owner as a bundle carries it: the id always, and
// the email when the user row came back with the project. A project whose
// owner's row is gone carries the id alone.
func projectOwner(project *entity.Project) *specmodel.ProjectOwner {
	if project.OwnerID == "" {
		return nil
	}
	owner := &specmodel.ProjectOwner{ID: project.OwnerID}
	if project.Owner != nil {
		owner.Email = project.Owner.Email
	}
	return owner
}

// assembleScope decrypts what the mode calls for and turns the settings into
// the block map a document carries.
func (s *service) assembleScope(
	settings []*entity.Setting,
	index *refIndex,
	mode specmodel.SecretsMode,
) (map[string]any, error) {
	for _, setting := range settings {
		if err := revealSettingSecrets(setting, mode); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return assembleSettings(settings, index, mode)
}

// indexSettings records the path each setting can be referenced by.
func indexSettings(index *refIndex, settings []*entity.Setting, scopePath string) {
	keys := specmodel.DeriveSettingKeys(settings)
	for _, setting := range settings {
		block := specmodel.SingletonBlockName(setting.Type)
		if block == "" {
			block = specmodel.CollectionBlockName(setting.Type) + "/" + keys[setting.ID]
		}
		index.addPath(setting.ID, scopePath+"/"+block)
	}
}

// specScopeName names a scope in a document.
//
// base.ObjectScopeGlobal is the empty string, which is fine as a database value
// and useless in a file somebody reads: "scope: """ says nothing. Every other
// scope already names itself.
func specScopeName(scope base.ObjectScopeType) string {
	if scope == base.ObjectScopeGlobal {
		return "global"
	}
	return string(scope)
}

func addFile(bundle *specmodel.Bundle, name string, doc any) error {
	content, err := marshalDoc(doc)
	if err != nil {
		return hperrors.Wrap(err)
	}
	bundle.Files[name] = content
	return nil
}

func sortedFilenames(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
