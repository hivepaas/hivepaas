package licenseserviceimpl

import (
	"context"
	"crypto/sha256"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// maxLastErrorLen bounds the last check's error as stored.
const maxLastErrorLen = 1000

// FetchRefresh asks with the stored key, which is what authenticates the
// request: the portal answers only for a key it issued, and only for the
// installation that key names. With no key there is nothing to ask with.
//
// It runs outside any transaction, as asking may take up to portalTimeout.
func (s *service) FetchRefresh(ctx context.Context, req *licenseservice.FetchRefreshReq) (
	*licenseservice.FetchRefreshResp, error) {
	settings, err := s.loadSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if settings.Key == "" {
		return &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshNone}, nil
	}
	cfg := s.cfg()
	if !req.Manual && cfg.License.RefreshDisabled {
		return &licenseservice.FetchRefreshResp{Result: base.LicenseRefreshDisabled}, nil
	}

	sentKeySum, checkedAt := sha256.Sum256([]byte(settings.Key)), s.timeNow().UTC()
	answer := s.ask(ctx, cfg, settings.Key)
	return &licenseservice.FetchRefreshResp{
		Result:     answer.result,
		SentKeySum: sentKeySum,
		NewKey:     answer.newKey,
		MovedAt:    answer.movedAt,
		CheckedAt:  checkedAt,
		Err:        answer.err,
	}, nil
}

// ApplyRefresh records every answer the portal gave, or failed to give, in the
// setting's Refresh. A check that asked nothing - no key, or the daily check
// switched off - is not recorded: it would rewrite the setting every day to say
// nothing happened.
func (s *service) ApplyRefresh(ctx context.Context, db database.Tx, fetched *licenseservice.FetchRefreshResp) (
	*licenseservice.ApplyRefreshResp, error) {
	switch fetched.Result {
	case base.LicenseRefreshNone, base.LicenseRefreshDisabled:
		return &licenseservice.ApplyRefreshResp{Result: fetched.Result}, nil
	case base.LicenseRefreshNewer, base.LicenseRefreshMoved, base.LicenseRefreshCurrent,
		base.LicenseRefreshUnknown, base.LicenseRefreshFailed:
	default:
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("a license refresh answered %q",
			fetched.Result)
	}

	setting, err := s.loadSetting(ctx, db, bunex.SelectFor("UPDATE OF setting"))
	if err != nil {
		return nil, err
	}
	settings, err := settingsOf(setting)
	if err != nil {
		return nil, err
	}
	// A key installed while the portal was asked is newer than anything the
	// answer, about the key it replaced, could install or say.
	if setting == nil || sha256.Sum256([]byte(settings.Key)) != fetched.SentKeySum {
		return &licenseservice.ApplyRefreshResp{Result: base.LicenseRefreshCurrent}, nil
	}

	applied := &licenseservice.ApplyRefreshResp{Result: fetched.Result, Err: fetched.Err}
	switch fetched.Result {
	case base.LicenseRefreshNewer:
		s.applyNewerKey(ctx, settings, fetched.NewKey, applied)
	case base.LicenseRefreshMoved:
		// The first answer that it moved says when; later ones say the same.
		if settings.MovedAt.IsZero() {
			settings.MovedAt = movedAt(fetched)
			applied.Changed = true
		}
	case base.LicenseRefreshCurrent, base.LicenseRefreshUnknown, base.LicenseRefreshFailed,
		base.LicenseRefreshNone, base.LicenseRefreshDisabled:
	}

	settings.Refresh.LastCheckedAt = fetched.CheckedAt
	settings.Refresh.LastResult = applied.Result
	settings.Refresh.LastError = errorText(applied.Err)
	setting, _, err = s.saveSettings(ctx, db, setting, settings)
	if err != nil {
		return nil, err
	}
	applied.Setting = setting
	if applied.Installed != nil {
		applied.Installed.Setting = setting
		s.dropVerification()
	}
	if applied.Changed && applied.State == nil {
		if applied.State, err = s.evaluate(ctx, settings); err != nil {
			return nil, err
		}
	}
	return applied, nil
}

// applyNewerKey installs the key the portal sent, refused as Install would
// refuse it: a portal is trusted with which key is newest, never with whether a
// key verifies or is for this installation. A key refused leaves the stored one
// in force, and the check failed.
func (s *service) applyNewerKey(ctx context.Context, settings *entity.LicenseSettings, newKey string,
	applied *licenseservice.ApplyRefreshResp) {
	newKey = license.NormalizeKeyText(newKey)
	if newKey == settings.Key {
		applied.Result = base.LicenseRefreshCurrent
		return
	}
	state, err := s.installable(ctx, newKey)
	if err != nil {
		applied.Result = base.LicenseRefreshFailed
		applied.Err = hperrors.Wrap(err).WithExtraDetail("The portal sent this newer key; the key stored is kept.")
		return
	}
	installKey(settings, newKey)
	applied.Changed = true
	applied.State = state
	applied.Installed = &licenseservice.InstallResp{State: state}
}

// movedAt is when the key moved, as the portal said; or, when it did not say,
// when it was asked. It is never later than that: the portal's clock running
// ahead of this one does not lengthen the grace a move leaves.
func movedAt(fetched *licenseservice.FetchRefreshResp) time.Time {
	if fetched.MovedAt.IsZero() || fetched.MovedAt.After(fetched.CheckedAt) {
		return fetched.CheckedAt
	}
	return fetched.MovedAt.UTC()
}

// errorText is the last check's error as System > License shows it: its message
// in the default language, never the key, which no error here holds.
func errorText(err error) string {
	if err == nil {
		return ""
	}
	info, _ := hperrors.ParseError(err, translation.GetDefaultLang())
	return truncate(info.Detail, maxLastErrorLen)
}

func (s *service) RefreshInfo(ctx context.Context) (*licenseservice.RefreshInfo, error) {
	settings, err := s.loadSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return &licenseservice.RefreshInfo{
		Enabled: !s.cfg().License.RefreshDisabled,
		Refresh: settings.Refresh,
	}, nil
}
