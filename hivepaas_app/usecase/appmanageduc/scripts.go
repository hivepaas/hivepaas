package appmanageduc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedscripts"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// scriptSettingTypes are the settings an app's scripts and artifacts are in.
var scriptSettingTypes = []base.SettingType{
	base.SettingTypeConfigFile, base.SettingTypeAppSettingMount, base.SettingTypeAppArtifacts,
}

// GetManagedScripts is what the installation's revision of a managed database's
// template offers it of what its engine owns, and whether it can be applied
// (docs/superpowers/specs/2026-10-21-managed-scripts-upgrade-design.md §1). A
// cluster's scripts are read on its app: a member's are upgraded with it.
func (uc *UC) GetManagedScripts(
	ctx context.Context,
	_ *basedto.Auth,
	req *appmanageddto.GetManagedScriptsReq,
) (*appmanageddto.GetManagedScriptsResp, error) {
	loaded, err := uc.loadManaged(ctx, uc.db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	if loaded.managed.MemberOf != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppManagedIsMember).
			WithParam("App", loaded.app.Name).ExposeParams("App")
	}
	view, err := uc.scriptsOf(ctx, uc.db, loaded)
	if err != nil {
		return nil, err
	}
	return &appmanageddto.GetManagedScriptsResp{Data: appmanageddto.TransformManagedScripts(view)}, nil
}

// scriptsOf reads the database's scripts, what its template's revision offers,
// and its members and verifier out of step.
func (uc *UC) scriptsOf(ctx context.Context, db database.IDB, loaded *loadedManaged) (
	*appmanageddto.ManagedScriptsView, error) {
	current, err := uc.scriptSetOf(ctx, db, loaded.app.ID)
	if err != nil {
		return nil, err
	}
	view := &appmanageddto.ManagedScriptsView{Current: current}
	view.Offer, view.CurrentRevision, view.AvailableRevision, err = uc.scriptsOffer(ctx, loaded, current)
	if err != nil {
		return nil, err
	}
	for _, member := range clusterMembers(loaded.managed) {
		set, err := uc.scriptSetOf(ctx, db, member.App.ID)
		if err != nil {
			return nil, err
		}
		if set.Fingerprint() != current.Fingerprint() {
			view.OutOfStep = append(view.OutOfStep, &appmanageddto.ManagedScriptsMemberResp{
				App: &basedto.ObjectIDResp{ID: member.App.ID}, Key: member.Key})
		}
	}
	return view, nil
}

// scriptsOffer is what the installation's revision of the app's template offers,
// with the revision the app's scripts came from and the revision offered. An
// app not made from a template, or whose template the revision lacks, is
// offered nothing.
func (uc *UC) scriptsOffer(ctx context.Context, loaded *loadedManaged, current *managedscripts.Set) (
	offer *managedscripts.Offer, currentRevision, availableRevision string, err error) {
	binding, err := uc.templateBinding(ctx, loaded.app)
	if errors.Is(err, hperrors.ErrNotFound) {
		return &managedscripts.Offer{State: managedscripts.StateUnavailable,
			Reason: "the database was not made from a template"}, "", "", nil
	}
	if err != nil {
		return nil, "", "", err
	}
	tmpl, err := uc.appTemplateService.Template(ctx, binding.Template)
	if errors.Is(err, hperrors.ErrAppTemplateNotFound) {
		return &managedscripts.Offer{State: managedscripts.StateUnavailable,
			Reason: "the templates no longer have " + binding.Template}, "", "", nil
	}
	if err != nil {
		return nil, "", "", hperrors.Wrap(err)
	}
	currentRevision = binding.Base.Revision
	if loaded.managed.Scripts != nil {
		currentRevision = loaded.managed.Scripts.Revision
	}
	return managedscripts.Evaluate(loaded.managed, tmpl.Template, binding.Version, current),
		currentRevision, tmpl.Revision, nil
}

// scriptSetOf is an app's scripts and artifacts.
func (uc *UC) scriptSetOf(ctx context.Context, db database.IDB, appID string) (*managedscripts.Set, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhereIn("setting.type IN (?)", scriptSettingTypes...),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	set, err := managedscripts.FromSettings(settings)
	return set, hperrors.Wrap(err)
}

// clusterMembers are the apps that run the cluster's scripts beside its own:
// its members, and its verifier, by app.
func clusterMembers(managed *entity.AppManagedSettings) []*entity.ManagedMember {
	state := managed.Database
	if state == nil {
		return nil
	}
	out := make([]*entity.ManagedMember, 0, len(state.Members)+1)
	out = append(out, state.Members...)
	if id := state.RestoreTest.VerifierApp.ID; id != "" {
		out = append(out, &entity.ManagedMember{App: entity.ObjectID{ID: id}, Key: "verifier"})
	}
	return out
}
