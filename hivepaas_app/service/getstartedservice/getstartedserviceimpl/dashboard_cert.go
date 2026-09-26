package getstartedserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

// dashboardRouting is the HivePaaS app and its routing settings: where the
// dashboard is served from.
type dashboardRouting struct {
	app     *entity.App
	routing *entity.AppRoutingSettings
}

func (s *service) loadDashboardRouting(ctx context.Context, db database.IDB) (*dashboardRouting, error) {
	app, err := s.hpAppService.LoadAppByKey(ctx, db, base.HivepaasAppKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	setting, err := s.settingRepo.GetSingle(ctx, db, app.GetObjectScope(), base.SettingTypeAppRouting, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	var routing *entity.AppRoutingSettings
	if setting != nil {
		if routing, err = setting.AsAppRoutingSettings(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &dashboardRouting{app: app, routing: routing}, nil
}

// dashboardDomain is the domain the dashboard is known by: the first enabled
// one it is served over HTTP at. TCP routes and TLS passed through are not
// what a browser opens the dashboard with.
func dashboardDomain(routing *entity.AppRoutingSettings) *entity.AppDomain {
	for _, domain := range routing.GetActiveDomains() {
		if domain.Protocol == base.NetworkProtocolTCP || domain.TLSPassthrough {
			continue
		}
		return domain
	}
	return nil
}

func (s *service) DashboardCert(ctx context.Context, db database.IDB) (*getstartedservice.Item, error) {
	dashboard, err := s.loadDashboardRouting(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	domain := dashboardDomain(dashboard.routing)
	if domain == nil {
		return &getstartedservice.Item{Status: getstartedservice.ItemStatusTodo}, nil
	}

	var attached *entity.SSLCert
	if domain.SSLCert.ID != "" {
		setting, err := s.settingRepo.GetByID(ctx, db, nil, base.SettingTypeSSLCert, domain.SSLCert.ID, true)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
		if setting != nil {
			if attached, err = setting.AsSSLCert(); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
	}

	pendingSetting, err := s.pendingCertSetting(ctx, db, dashboard.app, domain.Domain)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var pending *entity.SSLCert
	obtaining := false
	if pendingSetting != nil {
		if pending, err = pendingSetting.AsSSLCert(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		if obtaining, err = s.hasActiveObtainTask(ctx, db, pendingSetting.ID); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	item := dashboardCertItem(domain.Domain, attached, pending, obtaining, timeutil.NowUTC())
	return &item, nil
}

// dashboardCertItem decides the certificate item from what exists: a
// certificate attached to the domain, one being obtained for it, and whether a
// task is on that now.
//
// Only an attached certificate a browser trusts is done: the self-signed one a
// fresh install serves has the browser warn on every visit.
func dashboardCertItem(
	domain string,
	attached, pending *entity.SSLCert,
	obtaining bool,
	timeNow time.Time,
) getstartedservice.Item {
	item := getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: domain}
	switch {
	case attached != nil && attached.Certificate != "" && attached.CertType != base.SSLCertTypeSelfSigned &&
		(attached.ExpireAt.IsZero() || attached.ExpireAt.After(timeNow)):
		item.Status = getstartedservice.ItemStatusDone
	case obtaining:
		item.Status = getstartedservice.ItemStatusObtaining
	case pending != nil && pending.LastError != "":
		item.Status = getstartedservice.ItemStatusFailed
		item.Error = pending.LastError
	}
	return item
}

// pendingCertSetting is the certificate setting obtaining is going through for
// the domain, named for it, as domainservice names what it plans.
func (s *service) pendingCertSetting(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	domain string,
) (*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
		bunex.SelectWhere("setting.name = ?", domain),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return firstObtainable(settings), nil
}

// firstObtainable is the setting obtaining goes through. The installation's
// self-signed certificate can carry the same name - it is made for the root
// domain, which the dashboard's domain can be - and is never obtained.
func firstObtainable(settings []*entity.Setting) *entity.Setting {
	for _, setting := range settings {
		if setting.Kind != string(base.SSLCertTypeSelfSigned) {
			return setting
		}
	}
	return nil
}

// hasActiveObtainTask reports whether a task is still on the certificate: one
// waiting to run, running, or failed with a retry to come. A failed attempt
// leaves its task failed until the retry picks it up again, minutes later.
func (s *service) hasActiveObtainTask(ctx context.Context, db database.IDB, settingID string) (bool, error) {
	tasks, _, err := s.taskRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSSLObtain),
		bunex.SelectWhere("task.object_id = ?", settingID),
		bunex.SelectWhere("task.status IN (?)", bun.List([]base.TaskStatus{
			base.TaskStatusNotStarted, base.TaskStatusInProgress, base.TaskStatusFailed})),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return anyStillObtaining(tasks), nil
}

func anyStillObtaining(tasks []*entity.Task) bool {
	for _, task := range tasks {
		if task.IsNotStarted() || task.IsInProgress() || task.CanRetry() {
			return true
		}
	}
	return false
}

func (s *service) RequestDashboardCert(
	ctx context.Context,
	db database.IDB,
	ignoreRetryAfter bool,
) (*getstartedservice.CertRequest, error) {
	dashboard, err := s.loadDashboardRouting(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	domain := dashboardDomain(dashboard.routing)
	if domain == nil {
		return &getstartedservice.CertRequest{
			NotAsked: "the dashboard is served at no domain over HTTP: add one in the HivePaaS routing settings",
		}, nil
	}
	if domain.SSLCert.ID != "" {
		// The domain has a certificate someone chose: replacing it is theirs to do.
		return &getstartedservice.CertRequest{
			NotAsked: fmt.Sprintf("%s has a certificate attached already: change it in the HivePaaS "+
				"routing settings", domain.Domain),
		}, nil
	}
	// The dashboard's other names are asked for with it: one request, and every
	// address it answers at is covered.
	var wanted []string
	for _, active := range dashboard.routing.GetActiveDomains() {
		if active.SSLCert.ID == "" && active.Protocol != base.NetworkProtocolTCP && !active.TLSPassthrough {
			wanted = append(wanted, active.Domain)
		}
	}
	app := dashboard.app
	resp, err := s.domainService.EnsureCertsForDomains(ctx, db, &domainservice.EnsureCertsReq{
		Scope:            app.GetObjectScope(),
		ProjectID:        app.ProjectID,
		AppID:            app.ID,
		Domains:          wanted,
		IgnoreRetryAfter: ignoreRetryAfter,
		IgnoreSelfSigned: true,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	result := &getstartedservice.CertRequest{Tasks: resp.Tasks}
	if len(resp.Tasks) == 0 && len(resp.Obtaining) == 0 {
		result.NotAsked = notAskedReason(domain.Domain, resp)
	}
	return result, nil
}

// notAskedReason says why obtaining asked for nothing for the domain: what it
// skipped the name for, or the certificate it found covering it.
func notAskedReason(domain string, resp *domainservice.EnsureCertsResp) string {
	if reason := resp.Skipped[domain]; reason != "" {
		return fmt.Sprintf("no certificate was asked for %s: %s", domain, reason)
	}
	if cert := resp.Matched[domain]; cert != nil {
		return fmt.Sprintf("the certificate %q covers %s already: attach it to the domain in the HivePaaS "+
			"routing settings", cert.Name, domain)
	}
	return fmt.Sprintf("no certificate was asked for %s", domain)
}
