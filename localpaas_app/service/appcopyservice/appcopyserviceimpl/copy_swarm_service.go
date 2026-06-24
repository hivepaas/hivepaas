package appcopyserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/slugify"
	"github.com/localpaas/localpaas/localpaas_app/service/appservice"
)

func (s *service) copySwarmService(
	ctx context.Context,
	data *appCopyData,
) (err error) {
	targetApp := data.TargetApp
	srcSvcRes, err := s.dockerManager.ServiceInspect(ctx, data.SrcApp.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	srcSvc := &srcSvcRes.Service
	data.SrcService = srcSvc

	targetSvc := new(*srcSvc)
	data.TargetService = targetSvc

	targetSvc.ID = ""
	targetSvc.Spec.Name = targetApp.Key

	// Remove all env/config/secrets
	targetSvc.Spec.TaskTemplate.ContainerSpec.Env = nil
	targetSvc.Spec.TaskTemplate.ContainerSpec.Configs = nil
	targetSvc.Spec.TaskTemplate.ContainerSpec.Secrets = nil
	targetSvc.Spec.TaskTemplate.ContainerSpec.Hostname = targetApp.LocalKey

	// Update correct labels
	targetSvc.Spec.Labels[appservice.LabelAppNamespace] = data.TargetProject.Key
	targetSvc.Spec.Labels[appservice.LabelAppName] = targetApp.Name
	targetSvc.Spec.Labels[appservice.LabelAppEnv] = targetApp.Env

	// Update endpoints
	if targetSvc.Spec.EndpointSpec != nil {
		var ports []swarm.PortConfig
		for _, portConfig := range targetSvc.Spec.EndpointSpec.Ports {
			if portConfig.PublishMode == swarm.PortConfigPublishModeHost {
				continue
			}
			ports = append(ports, portConfig)
		}
		targetSvc.Spec.EndpointSpec.Ports = ports
	}

	// Update network attachments
	globalNetID, err := s.networkService.GetGlobalRoutingNetworkID(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	oldLocalNet, err := s.networkService.GetOrCreateProjectNetwork(ctx, data.SrcProject, data.SrcApp.Env)
	if err != nil {
		return apperrors.New(err)
	}
	newLocalNet, err := s.networkService.GetOrCreateProjectNetwork(ctx, data.TargetProject, data.TargetApp.Env)
	if err != nil {
		return apperrors.New(err)
	}
	var newNetAttachments []swarm.NetworkAttachmentConfig
	localNetAdded := false
	for _, net := range targetSvc.Spec.TaskTemplate.Networks {
		if net.Target == globalNetID || net.Target == base.NetworkGlobalRouting {
			newNetAttachments = append(newNetAttachments, net)
			continue
		}
		if oldLocalNet.ID != newLocalNet.ID && (net.Target == oldLocalNet.ID || net.Target == oldLocalNet.Name) {
			continue
		}
		if net.Target == newLocalNet.ID || net.Target == newLocalNet.Name {
			net.Aliases = []string{slugify.SlugifyAsKey(targetApp.Name)}
			newNetAttachments = append(newNetAttachments, net)
			localNetAdded = true
			continue
		}
		newNetAttachments = append(newNetAttachments, net)
	}
	if !localNetAdded { // Add local net
		newNetAttachments = append(newNetAttachments, swarm.NetworkAttachmentConfig{
			Target:  newLocalNet.ID,
			Aliases: []string{slugify.SlugifyAsKey(targetApp.Name)},
		})
	}
	targetSvc.Spec.TaskTemplate.Networks = newNetAttachments

	// TODO Update mounts

	err = data.CopyService(targetSvc, srcSvc)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) createSwarmService(
	ctx context.Context,
	data *appCopyData,
) (err error) {
	containerSpec := data.TargetService.Spec.TaskTemplate.ContainerSpec
	currImage := containerSpec.Image
	currCmd := containerSpec.Command
	currArgs := containerSpec.Args
	currDir := containerSpec.Dir
	currInit := containerSpec.Init
	containerSpec.Image = "busybox:latest"
	containerSpec.Command = []string{"sleep", "infinity"}
	containerSpec.Args = nil
	containerSpec.Dir = ""
	containerSpec.Init = new(true)

	defer func() {
		containerSpec.Image = currImage
		containerSpec.Command = currCmd
		containerSpec.Args = currArgs
		containerSpec.Dir = currDir
		containerSpec.Init = currInit
	}()

	// Create a service in docker for the app
	res, err := s.dockerManager.ServiceCreate(ctx, &data.TargetService.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	if res.ID == "" { // should never happen
		return apperrors.New(apperrors.ErrInfraInternal).
			WithNTParam("Error", "empty service ID returned")
	}
	data.TargetApp.ServiceID = res.ID
	data.TargetService.ID = res.ID
	return nil
}

func (s *service) applyFinalContainerSettings(
	ctx context.Context,
	data *appCopyData,
) error {
	inspect, err := s.dockerManager.ServiceInspect(ctx, data.TargetApp.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	service := &inspect.Service

	targetContainerSpec := data.TargetService.Spec.TaskTemplate.ContainerSpec
	containerSpec := service.Spec.TaskTemplate.ContainerSpec
	containerSpec.Image = targetContainerSpec.Image
	containerSpec.Command = targetContainerSpec.Command
	containerSpec.Args = targetContainerSpec.Args
	containerSpec.Dir = targetContainerSpec.Dir
	containerSpec.Init = targetContainerSpec.Init

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
