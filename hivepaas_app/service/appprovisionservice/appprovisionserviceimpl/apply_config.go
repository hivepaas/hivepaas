package appprovisionserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
)

const (
	routingApplyRetryMax   = 3
	routingApplyRetryDelay = 500 * time.Millisecond
)

func (s *service) ApplyAppConfiguration(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ApplyAppConfigurationReq,
) (*appprovisionservice.ApplyAppConfigurationResp, error) {
	app := req.App
	resp := &appprovisionservice.ApplyAppConfigurationResp{}

	// The environment comes first: the app kind makes shared variables such as
	// HIVEPAAS_PASSWORD, which the app's own variables and its secrets refer to.
	if err := s.applyEnvVars(ctx, db, app); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while applying environment variables")
	}
	// The app's files are its setting mounts': one update brings the service to
	// them.
	if err := s.settingMountService.Refresh(ctx, db, app); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while mounting the app's files")
	}
	if err := s.applyRouting(ctx, db, app, req.RefObjects, &resp.CertTasks); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while applying routing settings")
	}
	if err := s.applySchedJobs(ctx, db, app); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while scheduling the app's jobs")
	}
	return resp, nil
}

// applyEnvVars builds and applies the environment of every app in this app's
// scope, not only its own: a variable of another app may refer to this one.
func (s *service) applyEnvVars(ctx context.Context, db database.IDB, app *entity.App) error {
	// In a transaction: no nested transactions, and no concurrency.
	appEnvData, err := s.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, app.GetObjectScope(),
		false, nil, false, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	errMap := s.envVarService.ApplyEnvVarsForApps(ctx, db, appEnvData, false, false)
	for _, applyErr := range errMap {
		return hperrors.Wrap(applyErr)
	}
	return nil
}

// applyRouting writes the app's routing settings to traefik and to its service.
//
// It retries, which applying routing to an app that has been running does not
// have to: the service was created moments ago and swarm's own allocator is
// still writing to it, so an update carrying the version an inspect has just
// returned comes back as "update out of sequence" - about one create in three on
// a developer machine. Each attempt re-inspects the service and writes the same
// settings, so repeating it changes nothing beyond the version it carries.
func (s *service) applyRouting(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	refObjects *entity.RefObjects,
	tasks *[]*entity.Task,
) error {
	routingSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	if routingSetting == nil {
		return nil
	}
	routingSettings, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if refObjects == nil {
		refObjects = entity.NewRefObjects()
	}
	certTasks, err := s.attachCerts(ctx, db, app, routingSetting, routingSettings, refObjects)
	if err != nil {
		return hperrors.Wrap(err)
	}
	*tasks = append(*tasks, certTasks...)

	for attempt := range routingApplyRetryMax + 1 {
		if attempt > 0 {
			timer := time.NewTimer(routingApplyRetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return hperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}
		_, err = s.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
			App:             app,
			RoutingSettings: routingSettings,
			RefObjects:      refObjects,
		})
		if err == nil {
			return nil
		}
	}
	return hperrors.Wrap(err)
}

// attachCerts gives each domain that has no certificate of its own one that
// covers it - the certificate issued for that exact name, or a wildcard - and
// starts obtaining one where the system holds none.
//
// An app created with a domain has nobody to pick a certificate for it, and a
// wildcard is usually the reason the domain could be handed out at all. Choosing
// it here is what makes such an app served over TLS from its first deployment
// rather than after somebody opens its routing settings. A domain whose
// certificate has to be obtained is left as it is for now: the app is routed
// over plain HTTP, and the task attaches the certificate and routes it again
// when the authority has answered.
func (s *service) attachCerts(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	setting *entity.Setting,
	routing *entity.AppRoutingSettings,
	refObjects *entity.RefObjects,
) ([]*entity.Task, error) {
	var wanted []string
	for _, domain := range routing.GetActiveDomains() {
		if domain.SSLCert.ID == "" {
			wanted = append(wanted, domain.Domain)
		}
	}
	if len(wanted) == 0 {
		return nil, nil
	}

	certs, err := s.domainService.EnsureCertsForDomains(ctx, db, &domainservice.EnsureCertsReq{
		Scope:     app.GetObjectScope(),
		ProjectID: app.ProjectID,
		AppID:     app.ID,
		Domains:   wanted,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	attached := false
	for _, domain := range routing.GetActiveDomains() {
		cert := certs.Matched[domain.Domain]
		if domain.SSLCert.ID != "" || cert == nil {
			continue
		}
		domain.SSLCert = entity.ObjectID{ID: cert.ID}
		refObjects.RefSettings[cert.ID] = cert
		attached = true
	}
	if !attached {
		return certs.Tasks, nil
	}

	// The choice is written back, because it is the app's from now on: removing
	// the certificate has to find the apps using it, and the routing screen has
	// to show which one is serving the domain.
	if err = setting.SetData(routing); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.appService.PersistAppData(ctx, db, &appservice.PersistingAppData{
		UpsertingSettings: []*entity.Setting{setting},
	}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return certs.Tasks, nil
}

// applySchedJobs queues the first run of each job the app's settings schedule.
func (s *service) applySchedJobs(ctx context.Context, db database.IDB, app *entity.App) error {
	jobSettings := app.GetSettingsByType(base.SettingTypeSchedJob)
	if len(jobSettings) == 0 {
		return nil
	}
	tx, ok := db.(database.Tx)
	if !ok {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("db is not transaction")
	}
	return hperrors.Wrap(s.taskQueue.ScheduleTasksForSchedJobs(ctx, tx, jobSettings, false))
}
