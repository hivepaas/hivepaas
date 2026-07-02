package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
)

const (
	minAPIVersion = "1.52"
)

type Manager interface {
	// Config
	ConfigList(ctx context.Context, options ...ConfigListOption) (
		*client.ConfigListResult, error)
	ConfigInspect(ctx context.Context, configID string, options ...ConfigInspectOption) (
		*client.ConfigInspectResult, error)
	ConfigCreate(ctx context.Context, name string, data []byte, options ...ConfigCreateOption) (
		*client.ConfigCreateResult, error)
	ConfigRemove(ctx context.Context, configID string, options ...ConfigRemoveOption) (
		*client.ConfigRemoveResult, error)

	// Secrets
	SecretList(ctx context.Context, options ...SecretListOption) (
		*client.SecretListResult, error)
	SecretInspect(ctx context.Context, configID string, options ...SecretInspectOption) (
		*client.SecretInspectResult, error)
	SecretCreate(ctx context.Context, name string, data []byte, options ...SecretCreateOption) (
		*client.SecretCreateResult, error)
	SecretRemove(ctx context.Context, configID string, options ...SecretRemoveOption) (
		*client.SecretRemoveResult, error)

	// Containers
	ContainerList(ctx context.Context, options ...ContainerListOption) (
		*client.ContainerListResult, error)
	ServiceContainerList(ctx context.Context, serviceID string, options ...ContainerListOption) (
		*client.ContainerListResult, error)
	ContainerInspect(ctx context.Context, containerID string, options ...ContainerInspectOption) (
		*client.ContainerInspectResult, error)
	ContainerInspectMulti(ctx context.Context, containerIDs []string, options ...ContainerInspectOption) (
		map[string]*client.ContainerInspectResult, map[string]error)
	ContainerLogs(ctx context.Context, containerID string, options ...ContainerLogsOption) (
		client.ContainerLogsResult, error)

	ContainerRestart(ctx context.Context, containerID string, options ...ContainerRestartOption) (
		*client.ContainerRestartResult, error)
	ContainerRestartMulti(ctx context.Context, containerIDs []string, options ...ContainerRestartOption) (
		_ map[string]error)
	ContainerKill(ctx context.Context, containerID string, signal string, options ...ContainerKillOption) (
		*client.ContainerKillResult, error)
	ContainerKillMulti(ctx context.Context, containerIDs []string, signal string, options ...ContainerKillOption) (
		_ map[string]error)
	ContainerPrune(ctx context.Context, onlyObjectsOlderThan time.Duration, options ...ContainerPruneOption) (
		*client.ContainerPruneResult, error)

	ContainerExec(ctx context.Context, containerID string, options ...ExecCreateOption) (
		*client.ExecCreateResult, *client.ExecAttachResult, *client.ExecStartResult, error)
	ContainerExecWait(ctx context.Context, containerID string, options ...ExecCreateOption) (
		*client.ExecInspectResult, []*tasklog.LogFrame, error)
	ContainerExecResize(ctx context.Context, execID string, width, height uint) (
		*client.ExecResizeResult, error)
	ContainerExecInspect(ctx context.Context, execID string, options ...ExecInspectOption) (
		*client.ExecInspectResult, error)

	// Images
	ImageList(ctx context.Context, options ...ImageListOption) (
		*client.ImageListResult, error)
	ImagePull(ctx context.Context, name string, options ...ImagePullOption) (
		client.ImagePullResponse, error)
	ImagePush(ctx context.Context, name string, options ...ImagePushOption) (
		client.ImagePushResponse, error)
	ImageRemove(ctx context.Context, imageID string, options ...ImageRemoveOption) (
		*client.ImageRemoveResult, error)
	ImageInspect(ctx context.Context, imageID string, options ...ImageInspectOption) (
		*client.ImageInspectResult, error)
	ImagePrune(ctx context.Context, danglingOnly bool, onlyObjectsOlderThan time.Duration, options ...ImagePruneOption) (
		*client.ImagePruneResult, error)

	ImageBuild(ctx context.Context, buildContext io.Reader, options ...ImageBuildOption) (
		*client.ImageBuildResult, error)
	ImageBuildCancel(ctx context.Context, buildID string, options ...ImageBuildCancelOption) (
		*client.BuildCancelResult, error)
	BuildCachePrune(ctx context.Context, options ...BuildCachePruneOption) (
		*client.BuildCachePruneResult, error)

	// Networks
	NetworkList(ctx context.Context, options ...NetworkListOption) (
		*client.NetworkListResult, error)
	NetworkListByIDs(ctx context.Context, networkIDOrNames []string, options ...NetworkListOption) (
		*client.NetworkListResult, error)
	NetworkCreate(ctx context.Context, name string, options ...NetworkCreateOption) (
		*client.NetworkCreateResult, error)
	NetworkRemove(ctx context.Context, idOrName string, options ...NetworkRemoveOption) (
		*client.NetworkRemoveResult, error)
	NetworkInspect(ctx context.Context, name string, options ...NetworkInspectOption) (
		*client.NetworkInspectResult, error)
	NetworkExists(ctx context.Context, name string) bool
	NetworkPrune(ctx context.Context, onlyObjectsOlderThan time.Duration, options ...NetworkPruneOption) (
		*client.NetworkPruneResult, error)

	// Nodes
	NodeList(ctx context.Context, options ...NodeListOption) (
		*client.NodeListResult, error)
	NodeManagerList(ctx context.Context, options ...NodeListOption) (
		*client.NodeListResult, error)
	NodeInspect(ctx context.Context, nodeID string, options ...NodeInspectOption) (
		*client.NodeInspectResult, error)
	NodeUpdate(ctx context.Context, nodeID string, version *swarm.Version, spec *swarm.NodeSpec) (
		*client.NodeUpdateResult, error)
	NodeRemove(ctx context.Context, nodeID string, options ...NodeRemoveOption) (
		*client.NodeRemoveResult, error)
	NodeCurrentID(ctx context.Context) (string, error)

	// Registry
	RegistryLogin(ctx context.Context, options ...RegistryLoginOption) (*client.RegistryLoginResult, error)

	// Services
	ServiceList(ctx context.Context, options ...ServiceListOption) (
		*client.ServiceListResult, error)
	ServiceListByStack(ctx context.Context, namespace string, options ...ServiceListOption) (
		*client.ServiceListResult, error)
	ServiceGetByName(ctx context.Context, serviceName string, status bool) (
		*swarm.Service, error)
	ServiceInspect(ctx context.Context, serviceID string, options ...ServiceInspectOption) (
		*client.ServiceInspectResult, error)
	ServiceCreate(ctx context.Context, spec *swarm.ServiceSpec, options ...ServiceCreateOption) (
		*client.ServiceCreateResult, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, spec *swarm.ServiceSpec,
		options ...ServiceUpdateOption) (*client.ServiceUpdateResult, error)
	ServiceRollback(ctx context.Context, serviceID string, options ...ServiceUpdateOption) (
		*client.ServiceUpdateResult, error)
	ServiceForceUpdate(ctx context.Context, serviceID string) error
	ServiceRemove(ctx context.Context, serviceID string, options ...ServiceRemoveOption) (
		*client.ServiceRemoveResult, error)
	ServiceLogs(ctx context.Context, serviceID string, options ...ServiceLogsOption) (
		client.ServiceLogsResult, error)

	ServiceUpdateWait(ctx context.Context, serviceID string, inspectInterval time.Duration) (
		*swarm.Service, error)
	ServiceWaitUntilRunning(ctx context.Context, serviceID string, requireAllReplicas bool,
		requireRunningDuration time.Duration, checkInterval time.Duration) (bool, error)

	// Swarm
	SwarmInspect(ctx context.Context, options ...SwarmInspectOption) (
		*client.SwarmInspectResult, error)

	// System
	SystemInfo(ctx context.Context, options ...SystemInfoOption) (
		*client.SystemInfoResult, error)

	// Tasks
	TaskList(ctx context.Context, options ...TaskListOption) (
		*client.TaskListResult, error)
	ServiceTaskList(ctx context.Context, serviceID string, desiredStates []swarm.TaskState,
		options ...TaskListOption) (*client.TaskListResult, error)
	TaskInspect(ctx context.Context, taskID string, options ...TaskInspectOption) (
		*client.TaskInspectResult, error)
	TaskLogs(ctx context.Context, taskID string, options ...TaskLogsOption) (
		client.TaskLogsResult, error)
	ServiceTaskGetRunning(ctx context.Context, serviceID string, minRunningDuration time.Duration,
		maxRetry int, retryDelay time.Duration, ignoreNodeIDs []string) (
		running *swarm.Task, all *client.TaskListResult, err error)

	// Volumes
	VolumeList(ctx context.Context, options ...VolumeListOption) (
		*client.VolumeListResult, error)
	VolumeListByIDs(ctx context.Context, volumeIDOrNames []string, options ...VolumeListOption) (
		*client.VolumeListResult, error)

	VolumeCreate(ctx context.Context, options ...VolumeCreateOption) (
		*client.VolumeCreateResult, error)
	VolumeUpdate(ctx context.Context, volumeID string, version *swarm.Version, spec *volume.ClusterVolumeSpec) (
		*client.VolumeUpdateResult, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool, options ...VolumeRemoveOption) (
		*client.VolumeRemoveResult, error)
	VolumeInspect(ctx context.Context, volumeID string, options ...VolumeInspectOption) (
		*client.VolumeInspectResult, error)
	VolumePrune(ctx context.Context, anonymousOnly bool, options ...VolumePruneOption) (
		*client.VolumePruneResult, error)

	NewClientForNode(ctx context.Context, nodeID string) (Manager, error)

	Close() error
}

type manager struct {
	client *client.Client
}

func New() (Manager, error) {
	mgr := &manager{}
	c, err := client.New(
		client.WithAPIVersion(minAPIVersion),
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

func (m *manager) NewClientForNode(ctx context.Context, nodeID string) (Manager, error) {
	currNodeID, err := m.NodeCurrentID(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if currNodeID == nodeID {
		return m, nil
	}

	resp, err := m.TaskList(ctx, func(opts *client.TaskListOptions) {
		FilterAdd(&opts.Filters, "service", base.LocalpaasDockerProxyServiceName)
		FilterAdd(&opts.Filters, "node", nodeID)
		FilterAdd(&opts.Filters, "desired-state", string(swarm.TaskStateRunning))
	})
	if err != nil {
		return nil, apperrors.New(err)
	}
	if len(resp.Items) == 0 {
		return nil, apperrors.New(apperrors.ErrInfraNotFound).
			WithMsgLog("no running docker proxy task found on node %s", nodeID)
	}

	var targetIP string
	for _, netAttachment := range resp.Items[0].NetworksAttachments {
		if netAttachment.Network.Spec.Name == base.NetworkDockerProxy {
			if len(netAttachment.Addresses) > 0 {
				addr := netAttachment.Addresses[0]
				if addr.IsValid() {
					targetIP = addr.Addr().String()
					break
				}
			}
		}
	}

	if targetIP == "" {
		return nil, apperrors.New(apperrors.ErrInfraNotFound).
			WithMsgLog("docker proxy task on node %s is not connected to network %s", nodeID, base.NetworkDockerProxy)
	}

	remoteURL := fmt.Sprintf("tcp://%s:2375", targetIP)
	c, err := client.New(
		client.WithHost(remoteURL),
		client.WithAPIVersion(minAPIVersion),
	)
	if err != nil {
		return nil, apperrors.NewInfra(err)
	}

	return &manager{client: c}, nil
}
