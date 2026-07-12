package appsettingsuc

import (
	"context"
	"errors"
	"net/netip"
	"strings"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppNetworkSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
) (*appsettingsdto.UpdateAppNetworkSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppNetworkSettingsData{}
		err := uc.loadAppNetworkSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingAppData{}
		err = uc.prepareUpdatingAppNetworkSettings(req, data)
		if err != nil {
			return apperrors.New(err)
		}

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		err = uc.applyAppNetworkSettings(ctx, data)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.UpdateAppNetworkSettingsResp{}, nil
}

type updateAppNetworkSettingsData struct {
	App          *entity.App
	Service      *swarm.Service
	LocalNetwork *network.Inspect
}

func (uc *UC) loadAppNetworkSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
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
		return apperrors.New(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.New(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	// Loads project local network
	_, data.LocalNetwork, err = uc.networkService.GetOrCreateProjectNetwork(ctx, db, app.Project, app.Env)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}

	// Setting networks must be available in the project
	_, projectNets, err := uc.networkService.ListProjectNetworks(ctx, db, app.Project)
	if err != nil {
		return apperrors.New(err)
	}
	for _, newNet := range req.NetworkAttachments {
		if _, ok := projectNets[dockerhelper.ParseID(newNet.ID)]; !ok {
			return apperrors.New(apperrors.ErrProjectNetworkUnavailable).
				WithParam("Name", gofn.Coalesce(newNet.Name, newNet.ID))
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingAppNetworkSettings(
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) error {
	uc.prepareUpdatingAppNetworkAttachments(req, data)
	uc.prepareUpdatingAppHostsFileEntries(req, data)
	if err := uc.prepareUpdatingAppDNSConfig(req, data); err != nil {
		return apperrors.New(err)
	}
	uc.prepareUpdatingAppEndpointSpec(req, data)
	return nil
}

func (uc *UC) prepareUpdatingAppNetworkAttachments(
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) {
	service := data.Service
	localNetwork := data.LocalNetwork
	taskSpec := &service.Spec.TaskTemplate

	currNetworks := make(map[string]*swarm.NetworkAttachmentConfig, len(service.Spec.TaskTemplate.Networks))
	for i := range service.Spec.TaskTemplate.Networks {
		netAttachment := &service.Spec.TaskTemplate.Networks[i]
		currNetworks[netAttachment.Target] = netAttachment
	}

	taskSpec.Networks = make([]swarm.NetworkAttachmentConfig, 0, len(req.NetworkAttachments))
	for _, reqNet := range req.NetworkAttachments {
		reqNetID := dockerhelper.ParseID(reqNet.ID)
		// Skip if the net is already in the list
		if _, found := gofn.FindPtr(taskSpec.Networks, func(net *swarm.NetworkAttachmentConfig) bool {
			return net.Target == reqNetID
		}); found {
			continue
		}

		attachment := currNetworks[reqNetID]
		if attachment == nil {
			attachment = &swarm.NetworkAttachmentConfig{
				Target: reqNetID,
			}
		}
		attachment.Aliases = reqNet.Aliases
		// Special case: the network is the default project one
		if localNetwork != nil && (reqNetID == localNetwork.ID || reqNetID == localNetwork.Name) {
			defaultAlias := slugify.SlugifyAsKey(data.App.Name)
			if !gofn.Contain(attachment.Aliases, defaultAlias) {
				attachment.Aliases = append([]string{defaultAlias}, attachment.Aliases...)
			}
		}
		taskSpec.Networks = append(taskSpec.Networks, *attachment)
	}
}

func (uc *UC) prepareUpdatingAppHostsFileEntries(
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
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
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
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
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) error {
	service := data.Service
	if req.DNSConfig == nil {
		service.Spec.TaskTemplate.ContainerSpec.DNSConfig = nil
		return nil
	}
	containerSpec := service.Spec.TaskTemplate.ContainerSpec
	if containerSpec.DNSConfig == nil {
		containerSpec.DNSConfig = &swarm.DNSConfig{}
	}
	for _, addr := range req.DNSConfig.Nameservers {
		netAddr, err := netip.ParseAddr(addr)
		if err != nil {
			return apperrors.New(apperrors.ErrAddressInvalid).WithParam("Address", addr)
		}
		containerSpec.DNSConfig.Nameservers = append(containerSpec.DNSConfig.Nameservers, netAddr)
	}
	containerSpec.DNSConfig.Search = req.DNSConfig.Search
	containerSpec.DNSConfig.Options = req.DNSConfig.Options
	return nil
}

func (uc *UC) applyAppNetworkSettings(
	ctx context.Context,
	data *updateAppNetworkSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
