package traefikserviceimpl

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/traefikhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
)

func (s *service) OpenPorts(
	ctx context.Context,
	req *traefikservice.OpenPortReq,
) (resp *traefikservice.OpenPortResp, err error) {
	if req.Service == nil {
		svc, err := s.GetTraefikSwarmService(ctx)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		req.Service = svc
	}
	if req.Service == nil {
		return nil, nil
	}
	resp = &traefikservice.OpenPortResp{
		Service: req.Service,
	}

	applyFunc := func(_ int, svc *swarm.Service) (bool, error) {
		resp.Service = svc

		if svc.Spec.TaskTemplate.ContainerSpec == nil {
			svc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
		}
		if svc.Spec.EndpointSpec == nil {
			svc.Spec.EndpointSpec = &swarm.EndpointSpec{}
		}

		newArgs, argChanges := processPortConfigInArgs(svc.Spec.TaskTemplate.ContainerSpec.Args,
			req.OpenPorts, req.ClosePorts)

		newPortConfig, endpointSpecChanges := processPortConfigInEndpointSpec(svc.Spec.EndpointSpec,
			req.OpenPorts, req.ClosePorts)

		if !argChanges && !endpointSpecChanges {
			return false, nil
		}

		svc.Spec.TaskTemplate.ContainerSpec.Args = newArgs
		svc.Spec.EndpointSpec.Ports = newPortConfig

		if svc.Spec.UpdateConfig == nil {
			svc.Spec.UpdateConfig = &swarm.UpdateConfig{}
		}
		svc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
		svc.Spec.UpdateConfig.MaxFailureRatio = 0.5

		return true, nil
	}

	if req.SkipUpdatingServiceInDocker {
		_, err = applyFunc(0, req.Service)
	} else {
		err = s.dockerManager.ServiceUpdateFunc(ctx, req.Service.ID, req.Service, applyFunc,
			defaultServiceUpdateRetryMax, 0)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}

type parsedPort struct {
	portNum   uint32
	protocol  network.IPProtocol
	argKey    string
	argVal    string
	argString string
}

func (p parsedPort) key() string {
	return fmt.Sprintf("%d/%s", p.portNum, p.protocol)
}

func parsePortString(s string) (p parsedPort, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return p, false
	}
	s = strings.TrimPrefix(s, ":")
	portPart, protoPart, hasProto := strings.Cut(s, "/")
	portPart = strings.TrimSpace(portPart)
	portNum, err := strconv.ParseUint(portPart, 10, 32)
	if err != nil || portNum == 0 {
		return p, false
	}

	proto := network.TCP
	if hasProto && strings.EqualFold(protoPart, "udp") {
		proto = network.UDP
	}

	p.portNum = uint32(portNum)
	p.protocol = proto

	if proto == network.UDP {
		p.argKey = fmt.Sprintf("entrypoints.udp-svc-%d.address", portNum)
		p.argVal = fmt.Sprintf(":%d/udp", portNum)
		p.argString = fmt.Sprintf("--entrypoints.udp-svc-%d.address=:%d/udp", portNum, portNum)
	} else {
		p.argKey = fmt.Sprintf("entrypoints.tcp-svc-%d.address", portNum)
		p.argVal = fmt.Sprintf(":%d", portNum)
		p.argString = fmt.Sprintf("--entrypoints.tcp-svc-%d.address=:%d", portNum, portNum)
	}

	return p, true
}

func processPortConfigInArgs(
	currentArgs []string,
	openPorts, closePorts []string,
) (newArgs []string, hasChanges bool) {
	mapOpenPorts := make(map[string]parsedPort, len(openPorts))
	mapClosePorts := make(map[string]struct{}, len(closePorts))

	for _, openPort := range openPorts {
		if p, ok := parsePortString(openPort); ok {
			mapOpenPorts[p.key()] = p
		}
	}
	for _, closePort := range closePorts {
		if p, ok := parsePortString(closePort); ok {
			mapClosePorts[p.key()] = struct{}{}
		}
	}

	const tcpSvcPrefix = "entrypoints.tcp-svc-"
	const udpSvcPrefix = "entrypoints.udp-svc-"
	const suffix = ".address"

	newArgs = make([]string, 0, len(currentArgs)+len(openPorts))
	for _, arg := range currentArgs {
		key, _, valid := traefikhelper.ParseCommandArg(arg)
		if !valid || !strings.HasSuffix(key, suffix) {
			newArgs = append(newArgs, arg)
			continue
		}

		if strings.HasPrefix(key, tcpSvcPrefix) || strings.HasPrefix(key, udpSvcPrefix) {
			proto := network.TCP
			portStr := strings.TrimSuffix(strings.TrimPrefix(key, tcpSvcPrefix), suffix)
			if strings.HasPrefix(key, udpSvcPrefix) {
				proto = network.UDP
				portStr = strings.TrimSuffix(strings.TrimPrefix(key, udpSvcPrefix), suffix)
			}
			portNum, err := strconv.ParseUint(portStr, 10, 32)
			if err == nil {
				portKey := fmt.Sprintf("%d/%s", portNum, proto)
				if _, exists := mapClosePorts[portKey]; exists {
					hasChanges = true
					continue
				}
				delete(mapOpenPorts, portKey)
			}
		}
		newArgs = append(newArgs, arg)
	}

	for _, p := range mapOpenPorts {
		newArgs = append(newArgs, p.argString)
		hasChanges = true
	}

	return newArgs, hasChanges
}

func processPortConfigInEndpointSpec(
	endpointSpec *swarm.EndpointSpec,
	openPorts, closePorts []string,
) (newPortConfig []swarm.PortConfig, hasChanges bool) {
	mapOpenPorts := make(map[string]parsedPort, len(openPorts))
	mapClosePorts := make(map[string]struct{}, len(closePorts))

	for _, openPort := range openPorts {
		if p, ok := parsePortString(openPort); ok {
			mapOpenPorts[p.key()] = p
		}
	}
	for _, closePort := range closePorts {
		if p, ok := parsePortString(closePort); ok {
			mapClosePorts[p.key()] = struct{}{}
		}
	}

	var currentPorts []swarm.PortConfig
	if endpointSpec != nil {
		currentPorts = endpointSpec.Ports
	}

	newPortConfig = make([]swarm.PortConfig, 0, len(currentPorts)+len(openPorts))
	for _, portCfg := range currentPorts {
		proto := portCfg.Protocol
		if proto == "" {
			proto = network.TCP
		}
		portKey := fmt.Sprintf("%d/%s", portCfg.PublishedPort, proto)

		if _, exists := mapClosePorts[portKey]; exists {
			hasChanges = true
			continue
		}
		delete(mapOpenPorts, portKey)

		newPortConfig = append(newPortConfig, portCfg)
	}

	for _, p := range mapOpenPorts {
		newPortConfig = append(newPortConfig, swarm.PortConfig{
			Name:          fmt.Sprintf("%s-svc-%d", p.protocol, p.portNum),
			Protocol:      p.protocol,
			TargetPort:    p.portNum,
			PublishedPort: p.portNum,
			PublishMode:   swarm.PortConfigPublishModeIngress,
		})
		hasChanges = true
	}

	return newPortConfig, hasChanges
}
