package appdto

import (
	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetAppNetworkSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewGetAppNetworkSettingsReq() *GetAppNetworkSettingsReq {
	return &GetAppNetworkSettingsReq{}
}

func (req *GetAppNetworkSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppNetworkSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *NetworkSettingsResp `json:"data"`
}

type NetworkSettingsResp struct {
	NetworkAttachments []*NetworkAttachment `json:"networkAttachments"`
	HostsFileEntries   []*HostsFileEntry    `json:"hostsFileEntries"`
	DNSConfig          *DNSConfig           `json:"dnsConfig"`
	EndpointSpec       *EndpointSpec        `json:"endpointSpec"`

	UpdateVer int `json:"updateVer"`
}

type NetworkAttachment struct {
	ID      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type HostsFileEntry struct {
	Address   string   `json:"address,omitempty"`
	Hostnames []string `json:"hostnames,omitempty"`
}

type DNSConfig struct {
	Nameservers []string `json:"nameservers,omitempty"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type EndpointSpec struct {
	Mode  swarm.ResolutionMode `json:"mode,omitempty"`
	Ports []*PortConfig        `json:"ports,omitempty"`
}

type PortConfig struct {
	Target      uint32                      `json:"target,omitempty"`    // port inside the container
	Published   uint32                      `json:"published,omitempty"` // port on the swarm hosts
	Protocol    swarm.PortConfigProtocol    `json:"protocol,omitempty"`
	PublishMode swarm.PortConfigPublishMode `json:"publishMode,omitempty"`
}

func TransformNetworkSettings(
	service *swarm.Service,
	refObjects *InfraRefObjects,
) (resp *NetworkSettingsResp, err error) {
	spec := &service.Spec
	if refObjects == nil {
		refObjects = &InfraRefObjects{}
	}
	resp = &NetworkSettingsResp{
		UpdateVer: int(service.Version.Index), //nolint:gosec
	}

	resp.NetworkAttachments, err = TransformNetworkAttachments(spec.TaskTemplate.Networks, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.HostsFileEntries = TransformHostsFileEntries(spec.TaskTemplate.ContainerSpec.Hosts)
	resp.DNSConfig = TransformDNSConfig(spec.TaskTemplate.ContainerSpec.DNSConfig)
	resp.EndpointSpec = TransformEndpointSpec(spec.EndpointSpec)

	return resp, nil
}

func TransformNetworkAttachments(
	netAttachments []swarm.NetworkAttachmentConfig,
	refObjects *InfraRefObjects,
) (resp []*NetworkAttachment, err error) {
	resp = make([]*NetworkAttachment, 0, len(netAttachments))
	for _, netAttachment := range netAttachments {
		itemResp := &NetworkAttachment{
			ID: netAttachment.Target,
		}
		if net := refObjects.Networks[itemResp.ID]; net != nil {
			itemResp.Name = net.Name
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}

func TransformHostsFileEntries(hosts []string) (resp []*HostsFileEntry) {
	resp = make([]*HostsFileEntry, 0, len(hosts))
	for _, host := range hosts {
		parts := gofn.StringSplit(host, " ", "\"")
		resp = append(resp, &HostsFileEntry{
			Address:   parts[0],
			Hostnames: parts[1:],
		})
	}
	return resp
}

func TransformDNSConfig(config *swarm.DNSConfig) *DNSConfig {
	if config == nil {
		return nil
	}
	return &DNSConfig{
		Nameservers: config.Nameservers,
		Search:      config.Search,
		Options:     config.Options,
	}
}

func TransformEndpointSpec(endpointSpec *swarm.EndpointSpec) *EndpointSpec {
	if endpointSpec == nil {
		return nil
	}
	resp := &EndpointSpec{
		Mode:  endpointSpec.Mode,
		Ports: make([]*PortConfig, 0, len(endpointSpec.Ports)),
	}
	for _, port := range endpointSpec.Ports {
		resp.Ports = append(resp.Ports, &PortConfig{
			Target:      port.TargetPort,
			Published:   port.PublishedPort,
			Protocol:    port.Protocol,
			PublishMode: port.PublishMode,
		})
	}
	return resp
}
