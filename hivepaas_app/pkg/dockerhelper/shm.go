package dockerhelper

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func GetShmMount(taskSpec *swarm.TaskSpec) *mount.Mount {
	if taskSpec == nil || taskSpec.ContainerSpec == nil {
		return nil
	}
	for i := range taskSpec.ContainerSpec.Mounts {
		mnt := &taskSpec.ContainerSpec.Mounts[i]
		if mnt.Type != mount.TypeTmpfs || mnt.Target != "/dev/shm" {
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
			Target: "/dev/shm",
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
