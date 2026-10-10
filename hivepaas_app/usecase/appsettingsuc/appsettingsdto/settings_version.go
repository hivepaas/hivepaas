package appsettingsdto

import (
	"encoding/json"
	"hash/fnv"
	"slices"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// settingsVersion is the version a settings screen of an app is saved
// against: a hash of the part of the service the screen shows. The service's
// own version is no use for it - swarm moves it whenever the service is
// written: a rolling update marked complete a few seconds after its container
// runs, the autoscaler's replicas, another screen's save - and a screen nobody
// had changed was refused, "Mismatching update version". What the screen shows
// changed under it, it is refused still.
func settingsVersion(parts ...any) int {
	hash := fnv.New32a()
	_ = json.NewEncoder(hash).Encode(parts)
	return int(hash.Sum32())
}

// ResourceSettingsVersion is the version of the app's Resources screen.
func ResourceSettingsVersion(service *swarm.Service) int {
	task := &service.Spec.TaskTemplate
	return settingsVersion(TransformResourceReservations(task.Resources), TransformResourceLimits(task.Resources),
		TransformMemory(task), TransformCapabilities(task))
}

// ContainerSettingsVersion is the version of the app's Container Settings
// screen. The labels HivePaaS and docker manage are not of it: a save keeps
// them whatever it sends.
func ContainerSettingsVersion(service *swarm.Service) int {
	shown := TransformContainerSettingsBase(&service.Spec)
	if shown != nil {
		filtered := *shown
		filtered.ServiceLabels = dockerhelper.FilterOutRestrictedLabels(shown.ServiceLabels)
		filtered.ContainerLabels = dockerhelper.FilterOutRestrictedLabels(shown.ContainerLabels)
		shown = &filtered
	}
	return settingsVersion(shown)
}

// ServiceSettingsVersion is the version of the app's Availability screen: its
// mode, replicas and placement.
func ServiceSettingsVersion(service *swarm.Service) int {
	return settingsVersion(TransformServiceMode(&service.Spec),
		TransformServicePlacement(service.Spec.TaskTemplate.Placement))
}

// StorageSettingsVersion is the version of the app's Persistent Storage
// screen: its mounts, but the Docker API socket's, which is not chosen there.
func StorageSettingsVersion(service *swarm.Service) int {
	mounts := slices.DeleteFunc(slices.Clone(service.Spec.TaskTemplate.ContainerSpec.Mounts), func(m mount.Mount) bool {
		return dockerapiservice.IsSocketMount(&m)
	})
	return settingsVersion(mounts)
}

// NetworkSettingsVersion is the version of the app's Networks screen.
func NetworkSettingsVersion(service *swarm.Service) int {
	spec := &service.Spec
	containerSpec := spec.TaskTemplate.ContainerSpec
	return settingsVersion(spec.TaskTemplate.Networks, containerSpec.Hosts, containerSpec.DNSConfig, spec.EndpointSpec)
}
