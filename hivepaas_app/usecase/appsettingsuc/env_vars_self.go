package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice/envself"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// GetEnvSelfSuggestions suggests the variables an engine's image reads to set
// itself up, referring to the credentials the app's own kind settings publish.
func (uc *UC) GetEnvSelfSuggestions(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.GetEnvSelfSuggestionsReq,
) (*appsettingsdto.GetEnvSelfSuggestionsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	facts := &envself.Facts{Set: map[string]bool{}}
	if setting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind); setting != nil {
		kind, err := setting.AsAppKindSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		selfFacts(kind, facts)
	}

	appEngine := envself.EngineOf(facts.Engine)
	engine := appEngine
	if req.Engine != "" {
		engine = envself.EngineOf(req.Engine)
	}
	var suggestion *envself.Suggestion
	if engine != nil {
		suggestion = envself.Suggest(engine, facts)
	}
	return &appsettingsdto.GetEnvSelfSuggestionsResp{
		Data: appsettingsdto.TransformEnvSelfSuggestion(appEngine, suggestion),
	}, nil
}

// selfFacts are which of the credentials the kind settings publish have a
// value - read, never returned.
func selfFacts(kind *entity.AppKindSettings, facts *envself.Facts) {
	facts.Category, facts.Engine = kind.Category, kind.Engine
	switch {
	case kind.Category == base.AppCategoryDatabase && kind.Database != nil:
		db := kind.Database
		facts.Set[base.AppSystemEnvVarDatabaseName] = db.DbName != ""
		facts.Set[base.AppSystemEnvVarUser] = db.Username != ""
		facts.Set[base.AppSystemEnvVarPassword] = !db.Password.IsEmpty()
		facts.Set[base.AppSystemEnvVarRootPassword] = !db.RootPassword.IsEmpty()
	case kind.Category == base.AppCategoryCache && kind.Cache != nil:
		facts.Set[base.AppSystemEnvVarPassword] = !kind.Cache.Password.IsEmpty()
	}
}
