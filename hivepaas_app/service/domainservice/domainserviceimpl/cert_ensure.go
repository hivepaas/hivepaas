package domainserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
)

const (
	// sslObtainMaxRetry and sslObtainRetryDelay are how hard the task tries. A
	// certificate authority that refused once refuses the next attempt for the
	// same reason - a name that does not resolve here, a zone the DNS credentials
	// cannot write - so this is a short ladder, and the longer wait is
	// RetryAfter, which stops the next deployment from asking again.
	sslObtainMaxRetry   = 2
	sslObtainRetryDelay = timeutil.Duration(2 * time.Minute)
	// sslObtainRetryAfter is how long a failure is remembered. Let's Encrypt
	// counts five identical certificates a week; this keeps a redeploy loop from
	// spending that in an afternoon.
	sslObtainRetryAfter = 6 * time.Hour
)

// EnsureCertsForDomains gives every domain a certificate: the one the system
// already holds, or one it starts obtaining.
//
// Matching comes first and answers immediately - a wildcard covering the domain
// is the usual case and needs nothing. What is left is planned, recorded as
// certificate settings, and handed to a task: an authority takes seconds over an
// HTTP challenge and minutes over a DNS one, which is not something to hold a
// request open for. The caller gets back what can be attached now; the task
// attaches the rest when it has something to attach.
func (s *service) EnsureCertsForDomains(
	ctx context.Context,
	db database.IDB,
	req *domainservice.EnsureCertsReq,
) (*domainservice.EnsureCertsResp, error) {
	resp := &domainservice.EnsureCertsResp{Skipped: map[string]string{}}
	if len(req.Domains) == 0 {
		return resp, nil
	}

	matched, err := s.FindCertsForDomains(ctx, db, req.Scope, req.Domains)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if req.IgnoreSelfSigned {
		matched = withoutSelfSigned(matched)
	}
	resp.Matched = matched

	remaining := gofn.Filter(req.Domains, func(domain string) bool { return matched[domain] == nil })
	if len(remaining) == 0 {
		return resp, nil
	}

	certSettings, err := s.certPolicyFor(ctx, db, req.ProjectID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if certSettings == nil || !certSettings.AutoObtain {
		for _, domain := range remaining {
			resp.Skipped[domain] = "automatic certificates are off for this project"
		}
		return resp, nil
	}

	providers, err := s.certProvidersFor(ctx, db, req.Scope, certSettings.CertType)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// A certificate the installation signs itself is not validated by anybody, so
	// the names a public authority would refuse are worth obtaining here.
	publicAuthority := certTypeOf(certSettings) != base.SSLCertTypeSelfSigned
	plans, skipped := planCerts(remaining, providers.acme.ID != "", publicAuthority)
	for domain, reason := range skipped {
		resp.Skipped[domain] = reason
	}

	timeNow := timeutil.NowUTC()
	for _, plan := range plans {
		if err := s.ensurePlan(ctx, db, req, plan, certSettings, providers, timeNow, resp); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return resp, nil
}

// withoutSelfSigned is the matches a browser trusts.
func withoutSelfSigned(matched map[string]*entity.Setting) map[string]*entity.Setting {
	trusted := make(map[string]*entity.Setting, len(matched))
	for domain, setting := range matched {
		if setting.Kind != string(base.SSLCertTypeSelfSigned) {
			trusted[domain] = setting
		}
	}
	return trusted
}

// certProviders is what obtaining goes through: the account a certificate is
// bought or requested with, and the DNS credentials a wildcard needs.
type certProviders struct {
	ssl  entity.ObjectID
	acme entity.ObjectID
}

func (s *service) certProvidersFor(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	certType base.SSLCertType,
) (certProviders, error) {
	acme, err := s.firstSettingID(ctx, db, scope, base.SettingTypeAcmeDnsProvider)
	if err != nil {
		return certProviders{}, hperrors.Wrap(err)
	}
	ssl, err := s.defaultProviderFor(ctx, db, scope, certType)
	if err != nil {
		return certProviders{}, hperrors.Wrap(err)
	}
	return certProviders{ssl: ssl, acme: acme}, nil
}

// ensurePlan turns one planned certificate into a setting and a task, or says
// why it did not.
func (s *service) ensurePlan(
	ctx context.Context,
	db database.IDB,
	req *domainservice.EnsureCertsReq,
	plan *certPlan,
	certSettings *entity.DomainCertSettings,
	providers certProviders,
	timeNow time.Time,
	resp *domainservice.EnsureCertsResp,
) error {
	existing, err := s.certSettingNamed(ctx, db, req.Scope, plan.Name, req.IgnoreSelfSigned)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if existing != nil {
		// Something is already responsible for this name, so a second setting
		// would only mean a second request to the authority. What is left to
		// decide is whether to ask again through the setting there is.
		retry, reason := retryable(existing, timeNow, req.IgnoreRetryAfter)
		if !retry {
			for _, domain := range plan.Domains {
				resp.Skipped[domain] = reason
			}
			return nil
		}
		if err = s.markAttempt(ctx, db, existing, timeNow); err != nil {
			return hperrors.Wrap(err)
		}
		task, err := s.createObtainTask(ctx, db, existing, req.AppID, plan.Domains, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
		resp.Obtaining, resp.Tasks = append(resp.Obtaining, existing), append(resp.Tasks, task)
		return nil
	}

	setting, cert := certSettingFor(plan, certSettings, req.ProjectID, providers.ssl, providers.acme,
		gofn.Must(ulid.NewStringULID()), timeNow)
	if err = setting.SetData(cert); err != nil {
		return hperrors.Wrap(err)
	}
	if err = s.settingRepo.Insert(ctx, db, setting); err != nil {
		return hperrors.Wrap(err)
	}
	task, err := s.createObtainTask(ctx, db, setting, req.AppID, plan.Domains, timeNow)
	if err != nil {
		return hperrors.Wrap(err)
	}
	resp.Obtaining, resp.Tasks = append(resp.Obtaining, setting), append(resp.Tasks, task)
	return nil
}

// certPolicyFor is the certificate policy in force for a project: its own domain
// settings, or the installation's where the project has none.
func (s *service) certPolicyFor(
	ctx context.Context,
	db database.IDB,
	projectID string,
) (*entity.DomainCertSettings, error) {
	scope := entity.NewObjectScopeGlobal()
	if projectID != "" {
		scope = entity.NewObjectScopeProject(projectID)
	}
	setting, err := s.settingRepo.GetSingle(ctx, db, scope, base.SettingTypeDomainSettings, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting == nil {
		return nil, nil //nolint:nilnil // no settings is not an error: nothing is obtained
	}
	domainSettings, err := setting.AsDomainSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if domainSettings == nil {
		return nil, nil //nolint:nilnil
	}
	return domainSettings.CertSettings, nil
}

// firstSettingID is the first active setting of a type the scope can see - the
// DNS provider an installation configured, where there is one.
func (s *service) firstSettingID(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	typ base.SettingType,
) (entity.ObjectID, error) {
	settings, _, err := s.settingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", typ),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return entity.ObjectID{}, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return entity.ObjectID{}, nil
	}
	return entity.ObjectID{ID: settings[0].ID}, nil
}

// defaultProviderFor is the account a certificate is obtained through, when the
// installation has configured one. Let's Encrypt needs no account and is what
// an installation without a provider falls back to, so an empty answer is not a
// failure.
func (s *service) defaultProviderFor(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	certType base.SSLCertType,
) (entity.ObjectID, error) {
	settings, _, err := s.settingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLProvider),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return entity.ObjectID{}, hperrors.Wrap(err)
	}
	var fallback *entity.Setting
	for _, setting := range settings {
		if certType != "" && setting.Kind == string(certType) {
			return entity.ObjectID{ID: setting.ID}, nil
		}
		if setting.Default && fallback == nil {
			fallback = setting
		}
	}
	if fallback == nil {
		return entity.ObjectID{}, nil
	}
	return entity.ObjectID{ID: fallback.ID}, nil
}

// certSettingNamed finds the certificate setting already responsible for a name,
// whether it holds a certificate yet or not.
func (s *service) certSettingNamed(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	name string,
	ignoreSelfSigned bool,
) (*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, scope, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSSLCert),
		bunex.SelectWhere("setting.name = ?", name),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return firstNamed(settings, ignoreSelfSigned), nil
}

// firstNamed is the setting responsible for a name. The self-signed certificate
// an installation makes for its root domain is inherited everywhere under that
// name; a caller that wants one a browser trusts looks past it, or it would be
// told a certificate exists already.
func firstNamed(settings []*entity.Setting, ignoreSelfSigned bool) *entity.Setting {
	for _, setting := range settings {
		if ignoreSelfSigned && setting.Kind == string(base.SSLCertTypeSelfSigned) {
			continue
		}
		return setting
	}
	return nil
}

// retryable says whether a name that already has a certificate setting is worth
// asking an authority about again, and where it is not, what to tell the caller.
//
// The one case worth repeating is a failure that has waited out its RetryAfter:
// a DNS record that had not propagated yet, an app that was not reachable when
// the challenge came. An attempt still in flight is left to finish, and a
// setting that holds a certificate is renewal's business rather than this one's.
// ignoreWait is a person asking: it tries again at once, whatever the last
// attempt left, and the caller has made sure none is in flight.
func retryable(setting *entity.Setting, timeNow time.Time, ignoreWait bool) (retry bool, reason string) {
	cert, err := setting.AsSSLCert()
	if err != nil {
		return false, "a certificate named " + setting.Name + " already exists"
	}
	if cert.Certificate != "" {
		return false, "a certificate for " + setting.Name + " exists but cannot serve this domain"
	}
	if ignoreWait {
		return true, ""
	}
	if cert.LastError == "" {
		return false, "a certificate for " + setting.Name + " is already being obtained"
	}
	if timeNow.Before(cert.RetryAfter) {
		return false, "obtaining a certificate for " + setting.Name + " failed, waiting until " +
			cert.RetryAfter.Format(time.RFC3339) + ": " + cert.LastError
	}
	return true, ""
}

// markAttempt puts the wait back on before the task runs, so that two requests
// arriving together - a deployment and a routing change - schedule one attempt
// rather than two. The task rewrites it when it knows the outcome.
func (s *service) markAttempt(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	timeNow time.Time,
) error {
	cert := setting.MustAsSSLCert()
	cert.RetryAfter = timeNow.Add(sslObtainRetryAfter)
	if err := setting.SetData(cert); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdatedAt = timeNow
	if err := s.settingRepo.Update(ctx, db, setting,
		bunex.UpdateColumns("data", "updated_at")); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// createObtainTask writes the task that will ask an authority for the
// certificate. Handing it back unscheduled is deliberate: see EnsureCertsResp.
func (s *service) createObtainTask(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	appID string,
	domains []string,
	timeNow time.Time,
) (*entity.Task, error) {
	task := &entity.Task{
		ID:       gofn.Must(ulid.NewStringULID()),
		Scope:    base.ObjectScopeGlobal,
		ObjectID: setting.ID,
		Type:     base.TaskTypeSSLObtain,
		Status:   base.TaskStatusNotStarted,
		Config: entity.TaskConfig{
			Priority:   base.TaskPriorityDefault,
			MaxRetry:   sslObtainMaxRetry,
			RetryDelay: sslObtainRetryDelay,
		},
		Version:   entity.CurrentTaskVersion,
		RunAt:     timeNow,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := task.SetArgs(&entity.TaskSSLObtainArgs{
		SettingID: setting.ID,
		AppID:     appID,
		Domains:   domains,
	}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := s.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}
