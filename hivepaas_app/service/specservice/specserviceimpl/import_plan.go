package specserviceimpl

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func (s *service) ValidateImport(
	ctx context.Context,
	db database.IDB,
	req *specservice.ValidateImportReq,
) (*specservice.ValidateImportResp, error) {
	bundle, err := readBundle(req.Bundle, req.Passphrase)
	if err != nil {
		return nil, err
	}
	if mode := bundle.Manifest.SecretsMode; mode.RevealsSecrets() && req.AuthorizeSecrets != nil {
		if err = req.AuthorizeSecrets(ctx, mode); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	plan, err := s.planImport(ctx, db, req, bundle)
	if err != nil {
		return nil, err
	}
	return &specservice.ValidateImportResp{Plan: plan, SecretsMode: bundle.Manifest.SecretsMode}, nil
}

// planImport plans a bundle already read. It is the part tests reach directly,
// so they can change a document before planning it.
func (s *service) planImport(
	ctx context.Context,
	db database.IDB,
	req *specservice.ValidateImportReq,
	bundle *specmodel.ImportBundle,
) (*specmodel.ImportPlan, error) {
	if err := s.restrictToScope(ctx, db, req.Scope, bundle); err != nil {
		return nil, err
	}
	currentScope, err := s.currentScopeFor(ctx, db, req.Scope, bundle)
	if err != nil {
		return nil, err
	}
	current, err := s.currentState(ctx, db, currentScope, bundle.Manifest.SecretsMode)
	if err != nil {
		return nil, err
	}

	p := &planner{
		s: s, db: db, req: req, bundle: bundle, current: current,
		envOnly: req.Scope.ScopeType == base.ObjectScopeProjectEnv,
	}
	if err = p.plan(ctx); err != nil {
		return nil, err
	}
	manifest := bundle.Manifest
	return &specmodel.ImportPlan{
		Bundle: specmodel.BundleInfo{
			APIVersion: manifest.APIVersion, Scope: manifest.Scope, ExportedAt: manifest.ExportedAt,
			SourceAppVersion: manifest.SourceAppVersion, SecretsMode: manifest.SecretsMode,
		},
		Nodes:    p.nodes,
		Summary:  summarize(p.nodes),
		PlanHash: planHash(bundle.Digest, req, p.nodes),
	}, nil
}

// restrictToScope drops what the route may not write. A project or env route
// finds its own project in the bundle by id and then by key; a bundle without
// it is refused, since there would be nothing to import.
func (s *service) restrictToScope(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	bundle *specmodel.ImportBundle,
) error {
	if scope.ScopeType == base.ObjectScopeGlobal {
		return nil
	}
	project, err := s.projectRepo.GetByID(ctx, db, scope.ProjectID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	key := ""
	for bundleKey, doc := range bundle.Projects {
		if doc.ID != "" && doc.ID == project.ID || bundleKey == project.Key {
			key = bundleKey
		}
	}
	if key == "" {
		return hperrors.Wrap(hperrors.ErrSpecImportScopeNotInBundle)
	}
	bundle.Global = nil
	bundle.Projects = map[string]*specmodel.ProjectDoc{key: bundle.Projects[key]}
	bundle.Envs = map[string]map[string]*specmodel.EnvDoc{key: bundle.Envs[key]}

	if scope.ScopeType != base.ObjectScopeProjectEnv {
		return nil
	}
	_, envKey := projecthelper.ParseProjectEnvID(scope.ProjectEnvID)
	env, found := bundle.Envs[key][envKey]
	if !found {
		return hperrors.Wrap(hperrors.ErrSpecImportScopeNotInBundle)
	}
	bundle.Envs[key] = map[string]*specmodel.EnvDoc{envKey: env}
	// The env route writes the env and its apps, never its project's own record
	// or settings.
	bundle.Projects[key] = &specmodel.ProjectDoc{Project: key, ID: bundle.Projects[key].ID}
	return nil
}

// currentScopeFor is the scope this installation is exported at to compare
// with the bundle. It has to be as wide as the bundle's own: a reference to a
// global setting is a path in a global export and an external reference in any
// narrower one, and the two must not read as a difference. A global bundle is
// compared at global scope whatever the route; a narrower one at the route's
// scope, or at its project's when the route is global.
func (s *service) currentScopeFor(
	ctx context.Context,
	db database.IDB,
	route *entity.ObjectScope,
	bundle *specmodel.ImportBundle,
) (*entity.ObjectScope, error) {
	if bundle.Manifest.Scope == specScopeName(base.ObjectScopeGlobal) {
		return entity.NewObjectScopeGlobal(), nil
	}
	if route.ScopeType != base.ObjectScopeGlobal {
		return route, nil
	}
	for key, doc := range bundle.Projects {
		p := &planner{s: s, db: db}
		project, _, err := p.matchProject(ctx, doc, key)
		if err != nil {
			return nil, err
		}
		if project != nil {
			return entity.NewObjectScopeProject(project.ID), nil
		}
	}
	return route, nil
}

// currentState is this installation as export sees it, at the route's scope and
// in the bundle's secrets mode, parsed the way the upload was. Both sides are
// rendered by one piece of code, so a difference between them is a difference
// in configuration and not in how it was written down.
func (s *service) currentState(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	mode specmodel.SecretsMode,
) (*specmodel.ImportBundle, error) {
	built, err := s.buildBundle(ctx, db, &specservice.ExportReq{Scope: scope, SecretsMode: mode})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	manifest, err := marshalDoc(built.Manifest)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	files := maps.Clone(built.Files)
	files[manifestFilename] = manifest
	return parseBundleFiles(files)
}

// globalFilenameStem is the global node's path, as global.yaml is its file.
const globalFilenameStem = "global"

type planner struct {
	s       *service
	db      database.IDB
	req     *specservice.ValidateImportReq
	bundle  *specmodel.ImportBundle
	current *specmodel.ImportBundle
	nodes   []*specmodel.PlanNode
	// envOnly is an env route's plan: its project is matched, never planned.
	envOnly bool
	// missing names, by node path, the settings the target does not have at all.
	missing map[string][]string
}

func (p *planner) add(node *specmodel.PlanNode) *specmodel.PlanNode {
	node.Selected = p.req.Selection.Selects(node.Path)
	if node.Selected {
		node.SelectedBy = "user"
	}
	p.nodes = append(p.nodes, node)
	return node
}

func (p *planner) plan(ctx context.Context) error {
	if p.bundle.Global != nil {
		node := p.add(&specmodel.PlanNode{Path: globalFilenameStem, Kind: specmodel.NodeKindGlobal})
		var current map[string]any
		if p.current.Global != nil {
			current = p.current.Global.Settings
		}
		p.settingsNode(node, "", p.bundle.Global.Settings, current, true)
	}
	for _, key := range slices.Sorted(maps.Keys(p.bundle.Projects)) {
		if err := p.planProject(ctx, key); err != nil {
			return err
		}
	}
	for _, node := range p.nodes {
		p.applyExisting(node)
	}
	return nil
}

func (p *planner) planProject(ctx context.Context, key string) error {
	doc := p.bundle.Projects[key]
	path := "projects/" + key

	target, matchedBy, err := p.matchProject(ctx, doc, key)
	if err != nil {
		return err
	}
	if p.envOnly {
		for _, envKey := range slices.Sorted(maps.Keys(p.bundle.Envs[key])) {
			if err = p.planEnv(ctx, key, envKey, target); err != nil {
				return err
			}
		}
		return nil
	}
	node := p.add(&specmodel.PlanNode{Path: path, Kind: specmodel.NodeKindProject, Key: key, Name: doc.Name})
	var skip *specmodel.Issue
	switch {
	case target != nil && target.Key != key:
		skip = keyMismatch(path, target.Key)
	case target == nil:
		node.Action = specmodel.ActionCreate
		if taken, nameErr := p.s.projectRepo.GetByName(ctx, p.db, doc.Name); nameErr == nil && taken != nil {
			skip = &specmodel.Issue{Severity: specmodel.SeveritySkipped, Code: specmodel.CodeNameInUse, Path: path,
				Detail: map[string]any{"name": doc.Name}, Action: "not imported: another project has this name"}
		} else if nameErr != nil && !errors.Is(nameErr, hperrors.ErrNotFound) {
			return hperrors.Wrap(nameErr)
		}
	default:
		node.MatchedBy, node.TargetID = matchedBy, target.ID
		setChanges(node, p.projectChanges(doc, key))
	}

	var currentProject *specmodel.ProjectDoc
	if target != nil {
		currentProject = p.current.Projects[target.Key]
	}
	settingsNode := p.add(&specmodel.PlanNode{Path: path + "/settings", Kind: specmodel.NodeKindSettings, Key: key})
	var currentSettings map[string]any
	if currentProject != nil {
		currentSettings = currentProject.Settings
	}
	p.settingsNode(settingsNode, "", doc.Settings, currentSettings, target != nil)

	for _, envKey := range slices.Sorted(maps.Keys(p.bundle.Envs[key])) {
		if err = p.planEnv(ctx, key, envKey, target); err != nil {
			return err
		}
	}
	if skip != nil {
		p.skipBelow(path, *skip)
	}
	return nil
}

func (p *planner) matchProject(
	ctx context.Context, doc *specmodel.ProjectDoc, key string,
) (*entity.Project, specmodel.MatchedBy, error) {
	if doc.ID != "" {
		project, err := p.s.projectRepo.GetByID(ctx, p.db, doc.ID)
		if err == nil {
			return project, specmodel.MatchedByID, nil
		}
		if !errors.Is(err, hperrors.ErrNotFound) {
			return nil, "", hperrors.Wrap(err)
		}
	}
	project, err := p.s.projectRepo.GetByKey(ctx, p.db, key)
	if err == nil {
		return project, specmodel.MatchedByKey, nil
	}
	if !errors.Is(err, hperrors.ErrNotFound) {
		return nil, "", hperrors.Wrap(err)
	}
	return nil, "", nil
}

func (p *planner) projectChanges(doc *specmodel.ProjectDoc, key string) []string {
	current := p.current.Projects[key]
	if current == nil {
		return nil
	}
	var changes []string
	if doc.Name != current.Name {
		changes = append(changes, "name")
	}
	if doc.Note != current.Note {
		changes = append(changes, "note")
	}
	if doc.Owner != nil && (current.Owner == nil || doc.Owner.ID != current.Owner.ID) {
		changes = append(changes, "owner")
	}
	return changes
}

func (p *planner) planEnv(ctx context.Context, projectKey, envKey string, project *entity.Project) error {
	doc := p.bundle.Envs[projectKey][envKey]
	path := "projects/" + projectKey + "/envs/" + envKey
	node := p.add(&specmodel.PlanNode{Path: path, Kind: specmodel.NodeKindEnv, Key: envKey, Name: doc.Name})

	var env *entity.ProjectEnv
	if project != nil {
		found, err := p.s.projectEnvRepo.GetByKey(ctx, p.db, project.ID, envKey)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		env = found
	}
	var current *specmodel.EnvDoc
	if env == nil {
		node.Action = specmodel.ActionCreate
	} else {
		node.MatchedBy, node.TargetID = specmodel.MatchedByKey, env.ID
		current = p.current.Envs[project.Key][envKey]
		setChanges(node, envChanges(doc, current))
	}

	settingsNode := p.add(&specmodel.PlanNode{Path: path + "/settings", Kind: specmodel.NodeKindSettings, Key: envKey})
	var currentSettings map[string]any
	if current != nil {
		currentSettings = current.Settings
	}
	p.settingsNode(settingsNode, "", doc.Settings, currentSettings, env != nil)

	for _, appKey := range slices.Sorted(maps.Keys(doc.Apps)) {
		if err := p.planApp(ctx, path, appKey, doc.Apps[appKey], project, env, current); err != nil {
			return err
		}
	}
	return nil
}

func envChanges(doc, current *specmodel.EnvDoc) []string {
	if current == nil {
		return nil
	}
	var changes []string
	if doc.Name != current.Name {
		changes = append(changes, "name")
	}
	if doc.Color != current.Color {
		changes = append(changes, "color")
	}
	if doc.Index != current.Index {
		changes = append(changes, "index")
	}
	return changes
}

func (p *planner) planApp(
	ctx context.Context,
	envPath, key string,
	doc *specmodel.AppDoc,
	project *entity.Project,
	env *entity.ProjectEnv,
	currentEnv *specmodel.EnvDoc,
) error {
	path := envPath + "/apps/" + key
	node := p.add(&specmodel.PlanNode{Path: path, Kind: specmodel.NodeKindApp, Key: key, Name: doc.Name})

	app, matchedBy, err := p.matchApp(ctx, doc, key, project, env)
	if err != nil {
		return err
	}
	switch {
	case app != nil && (app.Key != key || env == nil || app.ProjectEnvID != env.ID):
		p.skipNode(node, *keyMismatch(path, app.Key))
		return nil
	case app == nil:
		node.Action = specmodel.ActionCreate
		node.Deploy = p.req.Options.DeployCreated && doc.Deployment != nil && doc.Deployment.Source != nil
		p.settingsChanges(node, "settings.", doc.Settings, nil)
		return nil
	}

	node.MatchedBy, node.TargetID = matchedBy, app.ID
	var current *specmodel.AppDoc
	if currentEnv != nil {
		current = currentEnv.Apps[key]
	}
	changes := deploymentChanges(doc.Deployment, current)
	var currentSettings map[string]any
	if current != nil {
		currentSettings = current.Settings
	}
	changes = append(changes, p.settingsChanges(node, "settings.", doc.Settings, currentSettings)...)
	setChanges(node, changes)
	node.Restart = app.ServiceID != "" && restarts(changes)
	node.Deploy = p.req.Options.DeployChangedSource && slices.Contains(changes, "deployment.source")
	return nil
}

func (p *planner) matchApp(
	ctx context.Context,
	doc *specmodel.AppDoc,
	key string,
	project *entity.Project,
	env *entity.ProjectEnv,
) (*entity.App, specmodel.MatchedBy, error) {
	if project == nil {
		return nil, "", nil
	}
	if doc.ID != "" {
		app, err := p.s.appRepo.GetByID(ctx, p.db, project.ID, doc.ID)
		if err == nil {
			return app, specmodel.MatchedByID, nil
		}
		if !errors.Is(err, hperrors.ErrNotFound) {
			return nil, "", hperrors.Wrap(err)
		}
	}
	if env == nil {
		return nil, "", nil
	}
	apps, _, err := p.s.appRepo.List(ctx, p.db, project.ID, nil,
		bunex.SelectWhere("app.project_env_id = ?", env.ID),
		bunex.SelectWhere("app.key = ?", key),
	)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	for _, app := range apps {
		if app.ProjectEnvID == env.ID && app.Key == key {
			return app, specmodel.MatchedByKey, nil
		}
	}
	return nil, "", nil
}

// restarts reports whether a change reaches a running app's service: every
// deployment block but the source, which only a deployment applies, and the
// settings ApplyAppConfiguration pushes to the service.
func restarts(changes []string) bool {
	for _, change := range changes {
		switch {
		case strings.HasPrefix(change, "deployment.") && change != "deployment.source":
			return true
		case change == "settings.envVars", change == "settings.routing",
			strings.HasPrefix(change, "settings.secrets"), strings.HasPrefix(change, "settings.configFiles"):
			return true
		}
	}
	return false
}
