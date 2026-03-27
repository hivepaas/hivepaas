package docker

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
)

type Manager interface {
	// Config
	ConfigList(ctx context.Context, options ...ConfigListOption) ([]swarm.Config, error)
	ConfigInspect(ctx context.Context, configID string) (*swarm.Config, error)
	ConfigCreate(ctx context.Context, name string, data []byte, options ...ConfigSpecOption) (
		*swarm.ConfigCreateResponse, error)

	// Containers
	ContainerList(ctx context.Context, options ...ContainerListOption) ([]container.Summary, error)
	ServiceContainerList(ctx context.Context, serviceID string, options ...ContainerListOption) (
		[]container.Summary, error)
	ServiceContainerGetActive(ctx context.Context, serviceID string, maxRetry int, retryDelay time.Duration) (
		active *container.Summary, all []container.Summary, err error)

	ContainerInspect(ctx context.Context, containerID string) (*container.InspectResponse, error)
	ContainerInspectMulti(ctx context.Context, containerIDs []string) (
		map[string]*container.InspectResponse, map[string]error)

	ContainerRestart(ctx context.Context, containerID string, options ...ContainerStopOption) error
	ContainerRestartMulti(ctx context.Context, containerIDs []string, options ...ContainerStopOption) map[string]error
	ContainerKill(ctx context.Context, containerID string, signal string) error
	ContainerKillMulti(ctx context.Context, containerIDs []string, signal string) map[string]error

	ContainerExec(ctx context.Context, containerID string, options *container.ExecOptions) (
		string, *types.HijackedResponse, error)
	ContainerExecWait(ctx context.Context, containerID string, options *container.ExecOptions) (
		*container.ExecInspect, []*applog.LogFrame, error)
	ContainerExecInspect(ctx context.Context, execID string) (*container.ExecInspect, error)

	ContainersPrune(ctx context.Context, onlyObjectsOlderThan time.Duration,
		options ...ContainersPruneOption) (*container.PruneReport, error)

	// Images
	ImageList(ctx context.Context, options ...ImageListOption) ([]image.Summary, error)
	ImageCreate(ctx context.Context, name string, options ...ImageCreateOption) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, imageID string, options ...ImageRemoveOption) ([]image.DeleteResponse, error)
	ImageInspect(ctx context.Context, imageID string) (*image.InspectResponse, error)
	ImagePull(ctx context.Context, refStr string, options ...ImagePullOption) (io.ReadCloser, error)
	ImagePush(ctx context.Context, imageTag string, options ...ImagePushOption) (io.ReadCloser, error)
	ImagesPrune(ctx context.Context, danglingOnly bool, onlyObjectsOlderThan time.Duration,
		options ...ImagesPruneOption) (*image.PruneReport, error)

	ImageBuild(ctx context.Context, buildContext io.Reader, options ...ImageBuildOption) (
		*build.ImageBuildResponse, error)
	ImageBuildCancel(ctx context.Context, buildID string) error

	// Networks
	NetworkList(ctx context.Context, options ...NetworkListOption) ([]network.Summary, error)
	NetworkCreate(ctx context.Context, name string, options ...NetworkCreateOption) (*network.CreateResponse, error)
	NetworkRemove(ctx context.Context, idOrName string) error
	NetworkInspect(ctx context.Context, name string, options ...NetworkInspectOption) (*network.Inspect, error)
	NetworkExists(ctx context.Context, name string) bool
	NetworksPrune(ctx context.Context, onlyObjectsOlderThan time.Duration,
		options ...NetworksPruneOption) (*network.PruneReport, error)

	// Nodes
	NodeList(ctx context.Context, options ...NodeListOption) ([]swarm.Node, error)
	NodeManagerList(ctx context.Context, options ...NodeListOption) ([]swarm.Node, error)
	NodeInspect(ctx context.Context, nodeID string) (*swarm.Node, []byte, error)
	NodeUpdate(ctx context.Context, nodeID string, version *swarm.Version, spec *swarm.NodeSpec) error
	NodeRemove(ctx context.Context, nodeID string, options ...NodeRemoveOption) error

	// Registry
	RegistryLogin(ctx context.Context, auth *registry.AuthConfig) (*registry.AuthenticateOKBody, error)

	// Services
	ServiceList(ctx context.Context, options ...ServiceListOption) ([]swarm.Service, error)
	ServiceListByStack(ctx context.Context, namespace string, options ...ServiceListOption) ([]swarm.Service, error)
	ServiceGetByName(ctx context.Context, serviceName string, options ...ServiceListOption) (*swarm.Service, error)

	ServiceCreate(ctx context.Context, service *swarm.ServiceSpec, options ...ServiceCreateOption) (
		*swarm.ServiceCreateResponse, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, service *swarm.ServiceSpec,
		options ...ServiceUpdateOption) (*swarm.ServiceUpdateResponse, error)
	ServiceForceUpdate(ctx context.Context, serviceID string) error
	ServiceRemove(ctx context.Context, serviceID string) error

	ServiceInspect(ctx context.Context, serviceID string, options ...ServiceInspectOption) (*swarm.Service, error)
	ServiceExists(ctx context.Context, serviceID string) bool
	ServiceLogs(ctx context.Context, serviceID string, options ...ContainerLogsOption) (io.ReadCloser, error)

	// Swarm
	SwarmInspect(ctx context.Context) (*swarm.Swarm, error)

	// System
	SystemInfo(ctx context.Context) (*system.Info, error)

	// Tasks
	TaskList(ctx context.Context, options ...TaskListOption) ([]swarm.Task, error)
	ServiceTaskList(ctx context.Context, serviceID string, options ...TaskListOption) ([]swarm.Task, error)

	// Volumes
	VolumeList(ctx context.Context, options ...VolumeListOption) (*volume.ListResponse, error)
	VolumeCreate(ctx context.Context, options *volume.CreateOptions) (*volume.Volume, error)
	VolumeUpdate(ctx context.Context, volumeID string, version *swarm.Version, options *volume.UpdateOptions) error
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	VolumeInspect(ctx context.Context, volumeID string) (*volume.Volume, []byte, error)
	VolumesPrune(ctx context.Context, anonymousOnly bool, options ...VolumesPruneOption) (*volume.PruneReport, error)

	Close() error
}

type manager struct {
	client *client.Client
}

func New() (Manager, error) {
	mgr := &manager{}
	c, err := client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}
	mgr.client = c
	return mgr, nil
}

func (m *manager) Close() error {
	return m.client.Close() //nolint:wrapcheck
}
