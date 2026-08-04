package volumeserviceimpl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

const (
	rsyncDefaultImage        = "alpine:latest"
	rsyncWaitPollInterval    = 2 * time.Second
	rsyncWaitTimeoutDuration = 30 * time.Minute
	fileModeFull             = 0777
)

func (s *service) Rsync(
	ctx context.Context,
	source, target *mount.Mount,
	options ...volumeservice.RsyncOption,
) error {
	opts := &volumeservice.RsyncOptions{
		Image: rsyncDefaultImage,
	}
	for _, opt := range options {
		opt(opts)
	}

	if source == nil || target == nil {
		return apperrors.Wrap(apperrors.ErrBadRequest).WithParam("Name", "source/target mount")
	}

	// Try Fast-Path: Direct Host Rsync if both source and target volumes are accessible on Host OS
	srcHostPath, srcDirect := s.getDirectHostPath(ctx, source, opts.SourceSubpath)
	destHostPath, destDirect := s.getDirectHostPath(ctx, target, opts.DestSubpath)

	if srcDirect && destDirect {
		return s.execDirectHostRsync(ctx, srcHostPath, destHostPath, opts)
	}

	// Fallback to container-based rsync Swarm Task
	return s.execContainerRsync(ctx, source, target, opts)
}

// getDirectHostPath checks if a volume or bind mount can be directly accessed on the Host OS.
// It inspects custom device paths (--opt device=/custom/path) as well as default Docker mountpoints.
func (s *service) getDirectHostPath(ctx context.Context, mnt *mount.Mount, subpath string) (string, bool) {
	if mnt == nil {
		return "", false
	}

	var hostPath string

	switch mnt.Type {
	case mount.TypeBind:
		// Bind Mount Source is directly a path on the Host Node
		hostPath = mnt.Source

	case mount.TypeVolume, mount.TypeCluster:
		// 1. Inspect Volume from Docker API
		volInspect, err := s.dockerManager.VolumeInspect(ctx, mnt.Source)
		if err != nil || volInspect == nil {
			return "", false
		}

		// 2. Check for custom device path if created with --opt device=/custom/path
		if devicePath, ok := volInspect.Volume.Options["device"]; ok && devicePath != "" {
			hostPath = devicePath
		} else {
			// Otherwise use default Docker-managed Mountpoint
			hostPath = volInspect.Volume.Mountpoint
		}

		if hostPath == "" {
			return "", false
		}

	case mount.TypeTmpfs, mount.TypeNamedPipe, mount.TypeImage:
	default:
		return "", false
	}

	// 3. Append Subpath if specified
	if subpath != "" {
		hostPath = filepath.Join(hostPath, subpath)
	}

	// 4. Check if hostPath exists on Host OS
	fileInfo, err := os.Stat(hostPath)
	if err != nil || !fileInfo.IsDir() {
		return "", false
	}

	// 5. Verify read/write permissions for HivePaaS Agent
	if !isWritableDir(hostPath) {
		return "", false
	}

	return hostPath, true
}

func isWritableDir(dir string) bool {
	testFile := filepath.Join(dir, fmt.Sprintf(".hivepaas_permcheck_%d", time.Now().UnixNano()))
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return true
}

// execDirectHostRsync executes rsync directly on the Host OS (Fast-Path)
func (s *service) execDirectHostRsync(
	ctx context.Context,
	srcPath, destPath string,
	opts *volumeservice.RsyncOptions,
) error {
	if !strings.HasSuffix(srcPath, "/") {
		srcPath += "/"
	}
	if !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}

	if opts.LogStore != nil {
		_ = opts.LogStore.Add(ctx, tasklog.NewOutFrame(
			fmt.Sprintf("Starting direct host volume rsync from %s to %s...", srcPath, destPath),
			tasklog.TsNow,
		))
	}

	// Ensure target directory exists on host
	if err := os.MkdirAll(destPath, fileModeFull); err != nil {
		return apperrors.Wrap(err)
	}

	args := []string{"-aHAX", "--info=progress2"}
	if opts.Delete {
		args = append(args, "--delete")
	}
	for _, exc := range opts.Exclude {
		if exc != "" {
			args = append(args, fmt.Sprintf("--exclude=%s", exc))
		}
	}
	args = append(args, srcPath, destPath)

	cmd := exec.CommandContext(ctx, "rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if opts.LogStore != nil {
			_ = opts.LogStore.Add(ctx, tasklog.NewErrFrame(
				fmt.Sprintf("Direct host rsync failed: %v, output: %s", err, string(output)),
				tasklog.TsNow,
			))
		}
		return apperrors.Wrap(err).WithParam("Output", string(output))
	}

	if opts.LogStore != nil {
		_ = opts.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Direct host volume rsync completed successfully.",
			tasklog.TsNow,
		))
	}
	return nil
}

// execContainerRsync executes rsync via a temporary helper Swarm Task using hivepaas_agent image (Fallback-Path)
//
//nolint:gocognit,funlen
func (s *service) execContainerRsync(
	ctx context.Context,
	source, dest *mount.Mount,
	opts *volumeservice.RsyncOptions,
) error {
	// Determine helper container image (use hivepaas_agent image if available)
	image := opts.Image
	if agentSvc, err := s.hpAppService.GetHpAgentSwarmService(ctx); err == nil && agentSvc != nil {
		if agentSvc.Spec.TaskTemplate.ContainerSpec != nil && agentSvc.Spec.TaskTemplate.ContainerSpec.Image != "" {
			image = agentSvc.Spec.TaskTemplate.ContainerSpec.Image
		}
	}

	// 1. Prepare mount specs for temporary helper container
	srcMnt := *source
	srcMnt.Target = "/from"
	srcMnt.ReadOnly = true

	destMnt := *dest
	destMnt.Target = "/to"
	destMnt.ReadOnly = false

	// Handle subpaths if specified
	srcPath := "/from"
	if opts.SourceSubpath != "" {
		srcPath = filepath.Join("/from", opts.SourceSubpath)
	}
	if !strings.HasSuffix(srcPath, "/") {
		srcPath += "/"
	}

	destPath := "/to"
	if opts.DestSubpath != "" {
		destPath = filepath.Join("/to", opts.DestSubpath)
	}
	if !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}

	// 2. Build rsync command (hivepaas_agent image already contains rsync)
	var cmdBuilder strings.Builder
	_, _ = fmt.Fprintf(&cmdBuilder, "mkdir -p %s && ", destPath)
	cmdBuilder.WriteString("rsync -aHAX --info=progress2 ")
	if opts.Delete {
		cmdBuilder.WriteString("--delete ")
	}
	for _, exc := range opts.Exclude {
		if exc != "" {
			_, _ = fmt.Fprintf(&cmdBuilder, "--exclude=%s ", exc)
		}
	}
	_, _ = fmt.Fprintf(&cmdBuilder, "%s %s", srcPath, destPath)

	rsyncCmd := cmdBuilder.String()

	if opts.LogStore != nil {
		_ = opts.LogStore.Add(ctx, tasklog.NewOutFrame(
			fmt.Sprintf("Starting container volume rsync using image '%s' from %s (%s) to %s (%s)...",
				image, source.Source, srcPath, dest.Source, destPath),
			tasklog.TsNow,
		))
	}

	// 3. Create a temporary Swarm Service for the rsync job
	serviceName := fmt.Sprintf("hivepaas-vol-rsync-%s", uuid.NewString()[:8])
	spec := &swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: serviceName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   image,
				Command: []string{"sh", "-c", rsyncCmd},
				Mounts:  []mount.Mount{srcMnt, destMnt},
			},
			RestartPolicy: &swarm.RestartPolicy{
				Condition: swarm.RestartPolicyConditionNone,
			},
		},
	}

	svcRes, err := s.dockerManager.ServiceCreate(ctx, spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	svcID := svcRes.ID

	// Ensure cleanup of the temporary Swarm service
	defer func() { //nolint:contextcheck
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second) //nolint:mnd
		defer cancel()
		_, _ = s.dockerManager.ServiceRemove(cleanupCtx, svcID)
	}()

	// 4. Poll for task completion
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, rsyncWaitTimeoutDuration)
	defer cancelTimeout()

	ticker := time.NewTicker(rsyncWaitPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return apperrors.Wrap(timeoutCtx.Err())
		case <-ticker.C:
			tasksRes, err := s.dockerManager.ServiceTaskList(timeoutCtx, svcID, nil)
			if err != nil || len(tasksRes.Items) == 0 {
				continue
			}

			// Get the latest task
			task := &tasksRes.Items[0]
			state := task.Status.State

			if state == swarm.TaskStateComplete {
				if opts.LogStore != nil {
					_ = opts.LogStore.Add(ctx, tasklog.NewOutFrame(
						"Volume rsync completed successfully.",
						tasklog.TsNow,
					))
				}
				return nil
			}

			if state == swarm.TaskStateFailed || state == swarm.TaskStateRejected {
				errMsg := task.Status.Err
				if errMsg == "" {
					errMsg = fmt.Sprintf("rsync task failed with state: %s", state)
				}
				if opts.LogStore != nil {
					_ = opts.LogStore.Add(ctx, tasklog.NewErrFrame(errMsg, tasklog.TsNow))
				}
				return apperrors.Wrap(apperrors.ErrInternal).WithParam("Reason", errMsg)
			}
		}
	}
}
