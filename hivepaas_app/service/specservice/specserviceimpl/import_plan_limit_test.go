package specserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// adding is the CheckReq an import makes: every limited kind it can add, each
// with its count.
func adding(apps, functions, signIns int) *licenseservice.CheckReq {
	return &licenseservice.CheckReq{Adding: map[base.LicenseLimit]int{
		base.LicenseLimitApps:            apps,
		base.LicenseLimitFunctions:       functions,
		base.LicenseLimitSignInProviders: signIns,
	}}
}

func licensesOf(svc *service) *fakeLicenseService {
	licenses, _ := svc.licenseService.(*fakeLicenseService)
	return licenses
}

// An app the import creates is checked on the import's transaction, which holds
// the plan-limit lock until the app is committed. The apps it updates are
// already counted.
func TestImportChecksTheAppsItCreatesOnItsTransaction(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, routingAt("worker.example.com"))
	bundle.Envs["project_a"]["dev"].Apps["worker"].Deployment.Source = imageSource("nginx:1.27")
	bundle.Envs["project_a"]["dev"].Apps["frontend"].Deployment = &specmodel.Deployment{Source: imageSource("web:2")}

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	licenses := licensesOf(svc)
	if assert.Len(t, licenses.reqs, 1) {
		assert.IsType(t, database.Tx{}, licenses.dbs[0])
		assert.Equal(t, adding(1, 0, 0), licenses.reqs[0])
	}
}

// A function in a bundle is created as a function, and counts as one.
func TestImportCountsAFunctionItCreatesAsAFunction(t *testing.T) {
	svc, bundle := planFixture(t)
	addReport(bundle, kindBody("function"), functionSourceBody("node24"), nil)

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	licenses := licensesOf(svc)
	if assert.Len(t, licenses.reqs, 1) {
		assert.Equal(t, adding(0, 1, 0), licenses.reqs[0])
	}
}

func TestImportOfAnUnchangedInstallationAddsNothing(t *testing.T) {
	svc, bundle := planFixture(t)

	apply(t, svc, bundle, applyReq(t, svc, bundle))

	licenses := licensesOf(svc)
	if assert.Len(t, licenses.reqs, 1) {
		assert.Equal(t, adding(0, 0, 0), licenses.reqs[0])
	}
}

// A refusal comes before anything of the import is persisted or provisioned.
func TestImportPastThePlanWritesNothing(t *testing.T) {
	svc, bundle := planFixture(t)
	addWorker(bundle, nil)
	bundle.Global.Settings["scripts"] = map[string]any{"build": map[string]any{"data": "make"}}
	req := applyReq(t, svc, bundle)
	licensesOf(svc).err = hperrors.Wrap(hperrors.ErrPlanLimitReached)

	_, err := svc.applyBundle(context.Background(), applyTx, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.Nil(t, svc.projectService.(*fakeProjectService).persisted, "nothing is persisted")
	assert.Empty(t, svc.appProvisionService.(*fakeProvisionService).reqs, "nothing is provisioned")
}

// The lock the plan is checked under lasts only as long as a transaction: an
// import on anything else is refused, whatever it adds.
func TestImportOutsideATransactionIsRefused(t *testing.T) {
	svc, bundle := planFixture(t)
	req := applyReq(t, svc, bundle)

	_, err := svc.applyBundle(context.Background(), nil, req, bundle)

	assert.ErrorIs(t, err, hperrors.ErrInternal)
	assert.Empty(t, licensesOf(svc).reqs)
}

// A paid sign-in provider the import creates active is one more; one whose
// secret the bundle left out is written pending, and adds nothing yet.
func TestImportCountsAPaidSignInItCreatesActive(t *testing.T) {
	for name, tc := range map[string]struct {
		kind   base.OAuthKind
		secret string
		want   int
	}{
		"paid, with its secret":    {base.OAuthKindGoogle, "s3cret", 1},
		"paid, its secret omitted": {base.OAuthKindGoogle, "", 0},
		"free":                     {base.OAuthKindGitlab, "s3cret", 0},
	} {
		svc, bundle := planFixture(t)
		bundle.Global.Settings["oauths"] = map[string]any{"sso": map[string]any{
			specmodel.SettingMetaKey: map[string]any{"kind": string(tc.kind), "version": entity.CurrentOAuthVersion},
			"clientId":               "client", "clientSecret": tc.secret,
		}}

		apply(t, svc, bundle, applyReq(t, svc, bundle))

		licenses := licensesOf(svc)
		if assert.Len(t, licenses.reqs, 1, name) {
			assert.Equal(t, adding(0, 0, tc.want), licenses.reqs[0], name)
		}
	}
}

// A sign-in provider is added when the import writes one active with a paid
// kind that was not that already: created so, enabled, or given a paid kind.
// Writing again one that was already is no addition: the providers that exist
// keep signing people in on every plan.
func TestImportAddsAPaidSignInOnlyWhereThereWasNone(t *testing.T) {
	oauth := func(kind base.OAuthKind, status base.SettingStatus) *entity.Setting {
		return &entity.Setting{ID: "oauth_1", Type: base.SettingTypeOAuth, Scope: base.ObjectScopeGlobal,
			Name: "sso", Kind: string(kind), Status: status}
	}
	active, disabled, pending := base.SettingStatusActive, base.SettingStatusDisabled, base.SettingStatusPending
	google, github := base.OAuthKindGoogle, base.OAuthKindGithub
	for name, tc := range map[string]struct {
		written, row *entity.Setting
		want         int
	}{
		"created active, paid":            {oauth(google, active), nil, 1},
		"created active, free":            {oauth(github, active), nil, 0},
		"created disabled, paid":          {oauth(google, disabled), nil, 0},
		"created pending, paid":           {oauth(google, pending), nil, 0},
		"rewritten as it was":             {oauth(google, active), oauth(google, active), 0},
		"enabled":                         {oauth(google, active), oauth(google, disabled), 1},
		"given a paid kind":               {oauth(google, active), oauth(github, active), 1},
		"disabled":                        {oauth(google, disabled), oauth(google, active), 0},
		"given a free kind":               {oauth(github, active), oauth(google, active), 0},
		"another type with the same kind": {withType(oauth(google, active), base.SettingTypeGithubApp), nil, 0},
		"created beside a disabled one":   {oauth(google, active), withID(oauth(google, disabled), "oauth_2"), 1},
	} {
		var rows []*entity.Setting
		if tc.row != nil {
			rows = append(rows, tc.row)
		}
		w := &writer{
			written: []*entity.Setting{tc.written},
			rows:    map[string]*scopeRows{"global": newScopeRows(rows)},
		}

		assert.Equal(t, tc.want, w.addedPaidSignIns(), name)
	}
}

func withType(setting *entity.Setting, typ base.SettingType) *entity.Setting {
	setting.Type = typ
	return setting
}

func withID(setting *entity.Setting, id string) *entity.Setting {
	setting.ID = id
	return setting
}

// withGlobalOAuth gives the fixture's installation a global sign-in provider,
// sso, the way the repository's SQL would return it.
func withGlobalOAuth(t *testing.T, svc *service, kind base.OAuthKind, status base.SettingStatus) {
	t.Helper()
	row := &entity.Setting{ID: "oauth_1", Type: base.SettingTypeOAuth, Scope: base.ObjectScopeGlobal,
		Name: "sso", Kind: string(kind), Status: status, Version: entity.CurrentOAuthVersion}
	assert.NoError(t, row.SetData(&entity.OAuth{ClientID: "client", ClientSecret: entity.NewEncryptedField("s3cret")}))
	loadOwned := svc.loadOwned
	svc.loadOwned = func(
		ctx context.Context, db database.IDB, scopes []base.ObjectScopeType, objectID string,
	) ([]*entity.Setting, error) {
		out, err := loadOwned(ctx, db, scopes, objectID)
		if containsScope(scopes, base.ObjectScopeGlobal) && objectID == "" {
			out = append(out, fresh(t, row))
		}
		return out, err
	}
}

// An import that writes this installation's provider again is matched to it: a
// paid one already active adds nothing, a paid one it enables adds one.
func TestImportOfAnExistingPaidSignInAddsOneOnlyWhenItEnablesIt(t *testing.T) {
	for name, tc := range map[string]struct {
		status base.SettingStatus
		want   int
	}{
		"active, its client changed": {base.SettingStatusActive, 0},
		"disabled, enabled":          {base.SettingStatusDisabled, 1},
	} {
		svc, bundle := planFixture(t)
		withGlobalOAuth(t, svc, base.OAuthKindGoogle, tc.status)
		bundle.Global.Settings["oauths"] = map[string]any{"sso": map[string]any{
			specmodel.SettingMetaKey: map[string]any{
				"kind": string(base.OAuthKindGoogle), "status": string(base.SettingStatusActive),
				"version": entity.CurrentOAuthVersion,
			},
			"clientId": "another-client",
		}}

		apply(t, svc, bundle, applyReq(t, svc, bundle))

		written := persistedSetting(t, svc, ofType(base.SettingTypeOAuth))
		assert.Equal(t, "oauth_1", written.ID, "matched to the provider: %s", name)
		licenses := licensesOf(svc)
		if assert.Len(t, licenses.reqs, 1, name) {
			assert.Equal(t, adding(0, 0, tc.want), licenses.reqs[0], name)
		}
	}
}
