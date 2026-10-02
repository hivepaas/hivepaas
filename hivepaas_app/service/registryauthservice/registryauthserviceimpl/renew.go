package registryauthserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// serviceUpdateRetryMax is how many times an update that lost to another
	// writer of the service is tried again.
	serviceUpdateRetryMax = 2
)

// Renew hands every service pulling with an ECR credential a token that lives
// past the next renewal. Only the registry auth sent with the update changes:
// the spec goes back as it was read, so Swarm leaves the tasks running. A
// credential or a service that fails is kept in the output and fails the run,
// after the others are done; the next run tries it again.
func (s *service) Renew(
	ctx context.Context,
	db database.IDB,
	req *registryauthservice.RenewReq,
) (*registryauthservice.RenewResp, error) {
	output := &entity.TaskRegistryAuthRenewalOutput{}
	resp := &registryauthservice.RenewResp{Output: output}
	interval := req.Interval
	if interval <= 0 {
		interval = entity.RegistryAuthRenewalIntervalDefault
	}

	auths, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'kind' = ?", base.RegistryAuthKindAWSECR),
		bunex.SelectWhereInIf(len(req.TargetAuths) > 0, "setting.id IN (?)", req.TargetAuths...),
	)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	if len(auths) == 0 {
		s.log(ctx, req, tasklog.NewOutFrame, "No Amazon ECR credential to renew.")
		return resp, nil
	}

	// A token for each credential, and its header
	headers := make(map[string]*authHeader, len(auths))
	for _, setting := range auths {
		header, err := s.renewalHeader(ctx, setting, interval)
		if err != nil {
			output.Failures = append(output.Failures, &entity.RegistryAuthRenewalFailure{
				Auth: setting.ID, Error: errorText(err)})
			s.log(ctx, req, tasklog.NewErrFrame, "No token for %s: %s", setting.Name, errorText(err))
			continue
		}
		headers[setting.ID] = header
		output.Auths = append(output.Auths, &entity.ObjectID{ID: setting.ID})
		s.log(ctx, req, tasklog.NewOutFrame, "Token for %s, until %s.", setting.Name,
			header.expiresAt.Format(time.RFC3339))
	}

	apps, err := s.appsPullingWith(ctx, db, headers)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	for _, app := range apps {
		header := headers[app.authID]
		err = s.updateService(ctx, app.app.ServiceID, header)
		switch {
		case errors.Is(err, errImageElsewhere):
			s.log(ctx, req, tasklog.NewWarnFrame, "%s skipped: its image is not in %s.", app.app.Name,
				header.address)
		case errors.Is(err, hperrors.ErrNotFound):
			s.log(ctx, req, tasklog.NewOutFrame, "%s skipped: it has no service.", app.app.Name)
		case err != nil:
			output.Failures = append(output.Failures, &entity.RegistryAuthRenewalFailure{
				Auth: app.authID, App: app.app.ID, Error: errorText(err)})
			s.log(ctx, req, tasklog.NewErrFrame, "%s not updated: %s", app.app.Name, errorText(err))
		default:
			output.Services = append(output.Services, &entity.RegistryAuthRenewalService{
				Auth: app.authID, App: app.app.ID, ServiceID: app.app.ServiceID})
			s.log(ctx, req, tasklog.NewOutFrame, "%s updated.", app.app.Name)
		}
	}

	if len(output.Failures) > 0 {
		return resp, hperrors.NewInfra(errRenewalFailed).
			WithExtraDetail("%d of the renewal's credentials and services failed: see the task's logs.",
				len(output.Failures))
	}
	return resp, nil
}

type authHeader struct {
	address   string
	header    string
	expiresAt time.Time
}

// renewalHeader is the credential's header with a token that lives the interval
// and the margin more, and when that token expires.
func (s *service) renewalHeader(
	ctx context.Context,
	setting *entity.Setting,
	interval time.Duration,
) (*authHeader, error) {
	auth, err := setting.AsRegistryAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	token, err := s.ecrTokenFor(ctx, setting, auth, interval)
	if err != nil {
		return nil, err
	}
	header, err := docker.GenerateAuthHeader(&registry.AuthConfig{
		Username: ecrUsername, Password: token.password, ServerAddress: auth.Address})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &authHeader{address: registryHost(auth.Address), header: header, expiresAt: token.expiresAt}, nil
}

type appPulling struct {
	app    *entity.App
	authID string
}

// appsPullingWith are the apps whose service pulls with one of the credentials:
// by their deployment's active method, the image's credential or that of the
// registry a build pushes to.
func (s *service) appsPullingWith(
	ctx context.Context,
	db database.IDB,
	headers map[string]*authHeader,
) ([]*appPulling, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	authIDs := make([]string, 0, len(headers))
	for id := range headers {
		authIDs = append(authIDs, id)
	}
	deployments, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhereOrGroup(
			bunex.SelectWhereIn("setting.data->'imageSource'->'registryAuth'->>'id' IN (?)", authIDs...),
			bunex.SelectWhereOr("setting.data->'repoSource'->'pushToRegistry'->>'id' IN (?)",
				bunex.List(authIDs)),
			bunex.SelectWhereOr("setting.data->'functionSource'->'pushToRegistry'->>'id' IN (?)",
				bunex.List(authIDs)),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	authOfApp := make(map[string]string, len(deployments))
	for _, setting := range deployments {
		deployment, err := setting.AsAppDeploymentSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if authID := deployment.ServiceRegistryAuthID(); headers[authID] != nil && setting.ObjectID != "" {
			authOfApp[setting.ObjectID] = authID
		}
	}
	if len(authOfApp) == 0 {
		return nil, nil
	}
	appIDs := make([]string, 0, len(authOfApp))
	for id := range authOfApp {
		appIDs = append(appIDs, id)
	}
	apps, err := s.appRepo.ListByIDs(ctx, db, "", appIDs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	res := make([]*appPulling, 0, len(apps))
	for _, app := range apps {
		if app.ServiceID == "" {
			continue
		}
		res = append(res, &appPulling{app: app, authID: authOfApp[app.ID]})
	}
	return res, nil
}

// errRenewalFailed is a run in which a credential or a service failed.
var errRenewalFailed = errors.New("the renewal of some credentials or services failed")

// errImageElsewhere is a service whose image is not in the credential's
// registry: its app's credential was changed and not yet deployed.
var errImageElsewhere = errors.New("the service's image is in another registry")

// updateService sends the service's spec back as it was read, with the
// credential's header. The image is checked first: a credential changed in the
// app's settings but not deployed would hand the service another registry's
// token.
func (s *service) updateService(ctx context.Context, serviceID string, header *authHeader) error {
	var elsewhere bool
	err := s.dockerManager.ServiceUpdateFunc(ctx, serviceID, nil,
		func(_ int, service *swarm.Service) (bool, error) {
			spec := service.Spec.TaskTemplate.ContainerSpec
			if spec == nil || !strings.HasPrefix(spec.Image, header.address+"/") {
				elsewhere = true
				return false, nil
			}
			return true, nil
		}, serviceUpdateRetryMax, 0,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = header.header
		})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if elsewhere {
		return errImageElsewhere
	}
	return nil
}

// registryHost is the address as an image's name starts with it.
func registryHost(address string) string {
	address = strings.TrimPrefix(strings.TrimPrefix(address, "https://"), "http://")
	return strings.TrimSuffix(address, "/")
}

func (s *service) log(
	ctx context.Context,
	req *registryauthservice.RenewReq,
	frame func(string, *time.Time) *tasklog.LogFrame,
	format string,
	args ...any,
) {
	if req.TaskExecData == nil || req.LogStore == nil {
		return
	}
	_ = req.LogStore.Add(ctx, frame(fmt.Sprintf(format, args...)+"\n", tasklog.TsNow))
}

// errorText is what a person is told of err: its detail when it has one.
func errorText(err error) string {
	info, _ := hperrors.ParseError(err, translation.LangEn)
	if info != nil && info.Detail != "" {
		return info.Detail
	}
	return err.Error()
}
