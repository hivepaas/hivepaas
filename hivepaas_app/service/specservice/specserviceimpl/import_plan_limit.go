package specserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// checkPlanLimits refuses an import that would take the installation past its
// plan: the apps and the functions it creates, and the paid sign-in providers
// it adds. A spec import and a project from a compose file both come here.
//
// It runs once the settings are built, and before anything is persisted or
// provisioned: an OAuth setting is written pending when the bundle left its
// secret out, and only one written active adds a provider.
func (w *writer) checkPlanLimits(ctx context.Context) error {
	// Apps and functions are counted under their plan-limit lock, held until the
	// transaction ends: on any other db the lock would be gone before what the
	// import creates is committed. Refused whatever the import adds, so that a
	// caller's mistake shows on the first import rather than the first to create.
	if _, isTx := w.p.db.(database.Tx); !isTx {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("spec import outside a transaction")
	}
	return hperrors.Wrap(w.p.s.licenseService.Check(ctx, w.p.db, w.planLimitReq()))
}

// planLimitReq is what the import adds of each limited kind.
func (w *writer) planLimitReq() *licenseservice.CheckReq {
	apps, functions := w.createdAppsByKind()
	return &licenseservice.CheckReq{Adding: map[base.LicenseLimit]int{
		base.LicenseLimitApps:            apps,
		base.LicenseLimitFunctions:       functions,
		base.LicenseLimitSignInProviders: w.addedPaidSignIns(),
	}}
}

// createdAppsByKind counts the apps the import creates, and of them the
// functions, which the plan limits apart. An app it updates is already counted,
// and an import never changes an app to or from a function.
func (w *writer) createdAppsByKind() (apps, functions int) {
	for _, node := range w.p.writingApps() {
		if node.Action != specmodel.ActionCreate {
			continue
		}
		if w.p.functionApps[node.Path] {
			functions++
			continue
		}
		apps++
	}
	return apps, functions
}

// addedPaidSignIns counts the paid sign-in providers the import adds: the OAuth
// settings it writes active with a paid kind that are new, or that were not
// active with a paid kind before. One written again as it was adds nothing: on
// every plan, the providers that exist keep signing people in.
func (w *writer) addedPaidSignIns() int {
	n := 0
	for _, setting := range w.written {
		if setting.Type != base.SettingTypeOAuth || !isActivePaidSignIn(setting) {
			continue
		}
		if row := w.replacedRow(setting); row == nil || !isActivePaidSignIn(row) {
			n++
		}
	}
	return n
}

// replacedRow is the setting of this installation a written one replaces, nil
// for one the import creates. A setting matched to a row is written under the
// row's id.
func (w *writer) replacedRow(setting *entity.Setting) *entity.Setting {
	for _, rows := range w.rows {
		if row := rows.byID[setting.ID]; row != nil {
			return row
		}
	}
	return nil
}

// isActivePaidSignIn reports whether an OAuth setting signs people in with a
// kind only a paid plan adds.
func isActivePaidSignIn(setting *entity.Setting) bool {
	return setting.Status == base.SettingStatusActive && license.IsPaidSignIn(base.OAuthKind(setting.Kind))
}
