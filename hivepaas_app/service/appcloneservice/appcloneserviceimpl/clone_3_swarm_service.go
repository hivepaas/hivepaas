package appcloneserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

//nolint:gocognit
func (s *service) cloneSwarmService(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	destApp, srcApp := data.DestApp, data.SrcApp
	srcSvcRes, err := s.dockerManager.ServiceInspect(ctx, srcApp.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	srcSvc := &srcSvcRes.Service
	data.SrcService = srcSvc

	destSvc := new(*srcSvc)
	data.DestService = destSvc

	destSvc.ID = ""
	destSvc.Spec.Name = destApp.GlobalKey

	if destSvc.Spec.TaskTemplate.ContainerSpec == nil {
		destSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}

	// Remove all env/config/secrets
	destSvc.Spec.TaskTemplate.ContainerSpec.Env = nil
	destSvc.Spec.TaskTemplate.ContainerSpec.Configs = nil
	destSvc.Spec.TaskTemplate.ContainerSpec.Secrets = nil
	destSvc.Spec.TaskTemplate.ContainerSpec.Hostname = destApp.Key

	// Set replicas
	targetReplicas := data.CloneSettings.TargetReplicas
	if targetReplicas < 0 && srcSvc.Spec.Mode.Replicated != nil && srcSvc.Spec.Mode.Replicated.Replicas != nil {
		targetReplicas = int(*srcSvc.Spec.Mode.Replicated.Replicas) //nolint:gosec
	}
	if targetReplicas < 0 {
		targetReplicas = 1
	}
	if destApp.Status != base.AppStatusActive {
		targetReplicas = 0
	}
	if destSvc.Spec.Mode.Replicated != nil {
		destSvc.Spec.Mode.Replicated.Replicas = new(uint64(targetReplicas))
	}

	// Update correct labels
	destSvc.Spec.Labels[appservice.LabelAppNamespace] = destApp.Project.Key
	destSvc.Spec.Labels[appservice.LabelAppKey] = destApp.Key
	destSvc.Spec.Labels[appservice.LabelAppName] = destApp.Name
	destSvc.Spec.Labels[appservice.LabelAppEnv] = destApp.ProjectEnv.Name

	// Update endpoints
	if destSvc.Spec.EndpointSpec != nil {
		var ports []swarm.PortConfig
		for _, portConfig := range destSvc.Spec.EndpointSpec.Ports {
			if portConfig.PublishMode == swarm.PortConfigPublishModeHost {
				continue
			}
			ports = append(ports, portConfig)
		}
		destSvc.Spec.EndpointSpec.Ports = ports
	}

	// Update network attachments
	globalNetID, err := s.networkService.GetGlobalRoutingNetworkID(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, oldLocalNet, err := s.networkService.GetOrCreateProjectNetwork(ctx, db, srcApp.Project,
		srcApp.ProjectEnv.Name)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, newLocalNet, err := s.networkService.GetOrCreateProjectNetwork(ctx, db, destApp.Project,
		data.DestApp.ProjectEnv.Name)
	if err != nil {
		return apperrors.Wrap(err)
	}
	var newNetAttachments []swarm.NetworkAttachmentConfig
	localNetAdded := false
	for _, net := range destSvc.Spec.TaskTemplate.Networks {
		if net.Target == globalNetID || net.Target == base.NetworkGlobalRouting {
			newNetAttachments = append(newNetAttachments, net)
			continue
		}
		if oldLocalNet.ID != newLocalNet.ID && (net.Target == oldLocalNet.ID || net.Target == oldLocalNet.Name) {
			continue
		}
		if net.Target == newLocalNet.ID || net.Target == newLocalNet.Name {
			net.Aliases = []string{destApp.Key}
			newNetAttachments = append(newNetAttachments, net)
			localNetAdded = true
			continue
		}
		newNetAttachments = append(newNetAttachments, net)
	}
	if !localNetAdded { // Add local net
		newNetAttachments = append(newNetAttachments, swarm.NetworkAttachmentConfig{
			Target:  newLocalNet.ID,
			Aliases: []string{destApp.Key},
		})
	}
	destSvc.Spec.TaskTemplate.Networks = newNetAttachments

	cloneFunc := data.OnCloneService
	if cloneFunc == nil {
		cloneFunc = func(destApp, srcApp *entity.App, destSvc, srcSvc *swarm.Service) error {
			return s.onCloneServiceDefault(destSvc, srcSvc, data)
		}
	}

	err = cloneFunc(destApp, srcApp, destSvc, srcSvc)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create a service in docker
	err = s.createSwarmService(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) onCloneServiceDefault(
	destSvc, _ *swarm.Service,
	data *appCloneData,
) error {
	settings := data.CloneSettings
	destSvcSpec := &destSvc.Spec
	containerSpec := destSvcSpec.TaskTemplate.ContainerSpec
	isDevEnv := config.Current.IsDevEnv()
	if !settings.CloneDeploymentSettings {
		containerSpec.Image = gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit)
		containerSpec.Command = nil
		containerSpec.Args = gofn.If(isDevEnv, nil, []string{"sleep", "infinity"})
		containerSpec.Dir = ""
	}

	return nil
}

func (s *service) createSwarmService(
	ctx context.Context,
	data *appCloneData,
) (err error) {
	// Create but not start the new service as the clone process doesn't complete
	createSpec := data.DestService.Spec
	createSpec.Mode = swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{Replicas: new(uint64(0))},
	}
	// Need to clone ContainerSpec before assigning temp values
	createSpec.TaskTemplate.ContainerSpec = new(*createSpec.TaskTemplate.ContainerSpec)
	createSpec.TaskTemplate.ContainerSpec.Image = "busybox:latest"
	createSpec.TaskTemplate.ContainerSpec.Command = nil
	createSpec.TaskTemplate.ContainerSpec.Args = []string{"sleep", "infinity"}
	createSpec.TaskTemplate.ContainerSpec.Dir = ""
	createSpec.TaskTemplate.ContainerSpec.Init = new(true)

	// Create a service in docker for the app
	res, err := s.dockerManager.ServiceCreate(ctx, &createSpec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if res.ID == "" { // should never happen
		return apperrors.Wrap(apperrors.ErrInfraInternal).
			WithNTParam("Error", "empty service ID returned")
	}
	data.DestApp.ServiceID = res.ID
	data.DestService.ID = res.ID
	return nil
}

func (s *service) applyFinalContainerSettings(
	ctx context.Context,
	data *appCloneData,
) (err error) {
	inspect, err := s.dockerManager.ServiceInspect(ctx, data.DestApp.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	destService := &inspect.Service
	updatingSpec := &destService.Spec

	// Restore service mode
	updatingSpec.Mode = data.DestService.Spec.Mode
	// Restore container spec
	updatingSpec.TaskTemplate.ContainerSpec = data.DestService.Spec.TaskTemplate.ContainerSpec

	_, err = s.dockerManager.ServiceUpdate(ctx, destService.ID, &destService.Version, updatingSpec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
