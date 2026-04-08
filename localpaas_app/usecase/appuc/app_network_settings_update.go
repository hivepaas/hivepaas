package appuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppNetworkSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppNetworkSettingsReq,
) (*appdto.UpdateAppNetworkSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppNetworkSettingsData{}
		err := uc.loadAppNetworkSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppNetworkSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppNetworkSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppNetworkSettingsResp{}, nil
}

type updateAppNetworkSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppNetworkSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppNetworkSettings(
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	uc.prepareUpdatingAppNetworkAttachments(req, data)
	uc.prepareUpdatingAppHostsFileEntries(req, data)
	uc.prepareUpdatingAppDNSConfig(req, data)
	uc.prepareUpdatingAppEndpointSpec(req, data)
}

func (uc *UC) prepareUpdatingAppNetworkAttachments(
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate

	currAttachments := make(map[string]*swarm.NetworkAttachmentConfig, len(taskSpec.Networks))
	for i := range taskSpec.Networks {
		netAttachment := &taskSpec.Networks[i]
		currAttachments[netAttachment.Target] = netAttachment
	}

	taskSpec.Networks = make([]swarm.NetworkAttachmentConfig, 0, len(req.NetworkAttachments))
	for _, reqNetAttachment := range req.NetworkAttachments {
		attachment := currAttachments[reqNetAttachment.ID]
		if attachment == nil {
			attachment = &swarm.NetworkAttachmentConfig{
				Target: reqNetAttachment.ID,
			}
		}
		attachment.Aliases = reqNetAttachment.Aliases
		taskSpec.Networks = append(taskSpec.Networks, *attachment)
	}
}

func (uc *UC) prepareUpdatingAppHostsFileEntries(
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	containerSpec.Hosts = make([]string, 0, len(req.HostsFileEntries))
	for _, host := range req.HostsFileEntries {
		s := append([]string{}, host.Address)
		s = append(s, host.Hostnames...)
		containerSpec.Hosts = append(containerSpec.Hosts, strings.Join(s, " "))
	}
}

func (uc *UC) prepareUpdatingAppEndpointSpec(
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	service := data.Service
	if req.EndpointSpec == nil {
		service.Spec.EndpointSpec = nil
		return
	}
	if service.Spec.EndpointSpec == nil {
		service.Spec.EndpointSpec = &swarm.EndpointSpec{}
	}
	endpointSpec := service.Spec.EndpointSpec
	endpointSpec.Mode = req.EndpointSpec.Mode
	endpointSpec.Ports = make([]swarm.PortConfig, 0, len(req.EndpointSpec.Ports))
	for _, port := range req.EndpointSpec.Ports {
		endpointSpec.Ports = append(endpointSpec.Ports, swarm.PortConfig{
			TargetPort:    port.Target,
			PublishedPort: port.Published,
			Protocol:      port.Protocol,
			PublishMode:   port.PublishMode,
		})
	}
}

func (uc *UC) prepareUpdatingAppDNSConfig(
	req *appdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	service := data.Service
	if req.DNSConfig == nil {
		service.Spec.TaskTemplate.ContainerSpec.DNSConfig = nil
		return
	}
	containerSpec := service.Spec.TaskTemplate.ContainerSpec
	if containerSpec.DNSConfig == nil {
		containerSpec.DNSConfig = &swarm.DNSConfig{}
	}
	containerSpec.DNSConfig.Nameservers = req.DNSConfig.Nameservers
	containerSpec.DNSConfig.Search = req.DNSConfig.Search
	containerSpec.DNSConfig.Options = req.DNSConfig.Options
}

func (uc *UC) applyAppNetworkSettings(
	ctx context.Context,
	data *updateAppNetworkSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
