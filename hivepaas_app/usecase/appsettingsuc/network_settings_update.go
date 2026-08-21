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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
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
			return apperrors.Wrap(err)
		}

		err = uc.applyAppNetworkSettings(ctx, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppNetworkSettingsResp{}, nil
}

type updateAppNetworkSettingsData struct {
	App            *entity.App
	Service        *swarm.Service
	LocalNetwork   *network.Inspect
	FinalNetworks  []swarm.NetworkAttachmentConfig
	NewNetworkReqs []*appsettingsdto.NetworkAttachment
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
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	// Loads project local network
	_, data.LocalNetwork, err = uc.networkService.GetOrCreateProjectNetwork(ctx, db, app.Project, app.ProjectEnv.Name)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}

	// Setting networks must be available in the project
	dbProjectNets, _, err := uc.networkService.ListProjectNetworks(ctx, db, app.Project)
	if err != nil {
		return apperrors.Wrap(err)
	}
	mapDbProjectNets := entityutil.SliceToIDMap(dbProjectNets)

	// Calculate current networks to distinguish new changes
	currNetworks := service.Spec.TaskTemplate.Networks
	mapCurrNetworkByID := make(map[string]*swarm.NetworkAttachmentConfig, len(currNetworks))
	for i := range currNetworks {
		net := &currNetworks[i]
		mapCurrNetworkByID[net.Target] = net
	}

	for _, netReq := range req.NetworkAttachments {
		// Network config has docker net ID (this can happen when admin configure networks not via HivePaaS)
		if existingNet := mapCurrNetworkByID[netReq.ID]; existingNet != nil {
			existingNet.Aliases = netReq.Aliases
			data.FinalNetworks = append(data.FinalNetworks, *existingNet)
			continue
		}
		// Regular network config using DB network ID
		dbNet := mapDbProjectNets[netReq.ID]
		if dbNet == nil {
			return apperrors.Wrap(apperrors.ErrProjectNetworkUnavailable).
				WithParam("Name", gofn.Coalesce(netReq.Name, netReq.ID))
		}
		data.FinalNetworks = append(data.FinalNetworks, swarm.NetworkAttachmentConfig{
			Target:  dbNet.MustAsClusterNetwork().RefID,
			Aliases: netReq.Aliases,
		})
	}

	return nil
}

func (uc *UC) prepareUpdatingAppNetworkSettings(
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) error {
	uc.prepareUpdatingAppNetworkAttachments(data)
	uc.prepareUpdatingAppHostsFileEntries(req, data)
	if err := uc.prepareUpdatingAppDNSConfig(req, data); err != nil {
		return apperrors.Wrap(err)
	}
	uc.prepareUpdatingAppEndpointSpec(req, data)
	return nil
}

func (uc *UC) prepareUpdatingAppNetworkAttachments(
	data *updateAppNetworkSettingsData,
) {
	localNetwork := data.LocalNetwork
	for i := range data.FinalNetworks {
		netAttach := &data.FinalNetworks[i]
		// Special case: the network is the default project one
		if localNetwork != nil && (netAttach.Target == localNetwork.ID || netAttach.Target == localNetwork.Name) {
			defaultAlias := slugify.SlugifyAsKey(data.App.Name)
			if !gofn.Contain(netAttach.Aliases, defaultAlias) {
				netAttach.Aliases = append([]string{defaultAlias}, netAttach.Aliases...)
			}
		}
	}

	data.Service.Spec.TaskTemplate.Networks = data.FinalNetworks
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
			return apperrors.Wrap(apperrors.ErrAddressInvalid).WithParam("Address", addr)
		}
		containerSpec.DNSConfig.Nameservers = append(containerSpec.DNSConfig.Nameservers, netAddr)
	}
	containerSpec.DNSConfig.Search = req.DNSConfig.Search
	containerSpec.DNSConfig.Options = req.DNSConfig.Options
	return nil
}

func (uc *UC) applyAppNetworkSettings(
	ctx context.Context,
	req *appsettingsdto.UpdateAppNetworkSettingsReq,
	data *updateAppNetworkSettingsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID,
		func(_ int, service *swarm.Service) error {
			data.Service = service
			return uc.prepareUpdatingAppNetworkSettings(req, data)
		}, defaultServiceRetryMax, 0, 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
