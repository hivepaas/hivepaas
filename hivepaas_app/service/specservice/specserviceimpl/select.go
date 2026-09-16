package specserviceimpl

import (
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// selectProjects drops the hivepaas project and orders the rest by key.
//
// That project holds HivePaaS's own stack - app, worker, db, redis, traefik,
// adminer, agent, updater, redis_ui - modeled as an ordinary project, every
// one of them with a live service id. Importing it elsewhere would redeploy
// HivePaaS's own database and proxy over the ones install.sh had just created.
// base.UnallowedProjectKeys already forbids anyone from creating another
// project with that key.
//
// It is not reported as an issue: leaving it out is the correct result, not
// something the operator needs to act on.
func selectProjects(projects []*entity.Project) []*entity.Project {
	kept := make([]*entity.Project, 0, len(projects))
	for _, project := range projects {
		if project.Key == base.HivepaasProjectKey {
			continue
		}
		kept = append(kept, project)
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Key < kept[j].Key })
	return kept
}

// selectApps drops preview apps and orders the rest by key.
//
// The only thing that sets ParentID is apppreviewservice - "Preview app must be
// a child app of the current" - so a child app is always a preview environment
// for a pull request, created when it opens and destroyed when it closes.
// Restoring one for a request that closed months ago is noise, and the preview
// machinery would remove it again anyway.
//
// This one is reported, because an operator looking for an app they can see in
// the dashboard should find out why it is not in the bundle.
func selectApps(apps []*entity.App, scopePath string, report *specmodel.Report) []*entity.App {
	kept := make([]*entity.App, 0, len(apps))
	for _, app := range apps {
		if app.ParentID != "" {
			report.Add(specmodel.Issue{
				Severity: specmodel.SeveritySkipped,
				Code:     specmodel.CodePreviewAppSkipped,
				Path:     scopePath + "/apps/" + app.Key,
				Detail:   map[string]any{"parentId": app.ParentID},
				Action:   "not exported; preview apps are recreated by their own lifecycle",
			})
			continue
		}
		kept = append(kept, app)
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].Key < kept[j].Key })
	return kept
}

// selectSettings applies each type's export policy and records every refusal.
//
// Ordering is by type, then created_at, then id - the same ordering
// DeriveSettingKeys uses, so a key and the setting it names agree on which came
// first.
func selectSettings(
	settings []*entity.Setting,
	scopePath string,
	report *specmodel.Report,
) []*entity.Setting {
	kept := make([]*entity.Setting, 0, len(settings))
	for _, setting := range settings {
		decision := entity.SpecExportDecision(setting)
		if decision.Export {
			kept = append(kept, setting)
			continue
		}

		code := specmodel.CodeTypeSkipped
		if entity.SpecPolicyFor(setting.Type) == nil {
			code = specmodel.CodeTypeUnclassified
		}
		report.Add(specmodel.Issue{
			Severity: specmodel.SeveritySkipped,
			Code:     code,
			Path:     scopePath + "/" + string(setting.Type) + "/" + setting.Name,
			Detail:   map[string]any{"type": string(setting.Type), "id": setting.ID},
			Action:   decision.Reason,
		})
	}

	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].Type != kept[j].Type {
			return kept[i].Type < kept[j].Type
		}
		if !kept[i].CreatedAt.Equal(kept[j].CreatedAt) {
			return kept[i].CreatedAt.Before(kept[j].CreatedAt)
		}
		return kept[i].ID < kept[j].ID
	})
	return kept
}
