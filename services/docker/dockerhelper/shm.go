package dockerhelper

import (
	"slices"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// shmTarget is where a container's shared memory is mounted.
const shmTarget = "/dev/shm"

func GetShmMount(taskSpec *swarm.TaskSpec) *mount.Mount {
	if taskSpec == nil || taskSpec.ContainerSpec == nil {
		return nil
	}
	for i := range taskSpec.ContainerSpec.Mounts {
		mnt := &taskSpec.ContainerSpec.Mounts[i]
		if mnt.Type != mount.TypeTmpfs || mnt.Target != shmTarget {
			continue
		}
		return mnt
	}
	return nil
}

func SetShmSize(taskSpec *swarm.TaskSpec, size int64) *mount.Mount {
	if taskSpec == nil || taskSpec.ContainerSpec == nil {
		return nil
	}
	shmMount := GetShmMount(taskSpec)
	if shmMount == nil {
		shmMount = &mount.Mount{
			Type:   mount.TypeTmpfs,
			Target: shmTarget,
			TmpfsOptions: &mount.TmpfsOptions{
				SizeBytes: size,
				Mode:      base.DirModeDefault,
			},
		}
		taskSpec.ContainerSpec.Mounts = append(taskSpec.ContainerSpec.Mounts, *shmMount)
	} else {
		shmMount.TmpfsOptions.SizeBytes = size
	}
	return nil
}

// RemoveShmMount takes the task's own /dev/shm away: its container has the
// size docker gives it.
func RemoveShmMount(taskSpec *swarm.TaskSpec) {
	if taskSpec == nil || taskSpec.ContainerSpec == nil {
		return
	}
	taskSpec.ContainerSpec.Mounts = slices.DeleteFunc(taskSpec.ContainerSpec.Mounts, func(mnt mount.Mount) bool {
		return mnt.Type == mount.TypeTmpfs && mnt.Target == shmTarget
	})
}
