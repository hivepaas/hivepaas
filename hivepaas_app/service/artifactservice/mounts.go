package artifactservice

import (
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// CacheSubdir is where every node's agent keeps artifacts under HivePaaS's host
// directory.
const CacheSubdir = "cache/artifacts/sha256"

// CacheHostDir is the artifacts cache as the host sees it, the same on every
// node: the stack gives every agent the one HP_STORAGE_HOST_DIR. It is "" when
// no host directory is configured, and nothing is then an artifact.
func CacheHostDir() string {
	cfg := config.Current()
	if cfg == nil || cfg.Storage.HostDir == "" {
		return ""
	}
	return filepath.Join(cfg.Storage.HostDir, CacheSubdir)
}

// Mount binds a cached artifact into a container, read-only. It never creates
// its mount point: a missing file has to stop the task, not become a directory
// that the app then runs into.
func Mount(hostPath, target string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: hostPath, Target: target, ReadOnly: true}
}

// IsArtifactMount reports a mount that artifacts give an app: a bind of a file
// in the artifacts cache. The app's own mounts never bind there.
func IsArtifactMount(m *mount.Mount) bool {
	cacheDir := CacheHostDir()
	return cacheDir != "" && m.Type == mount.TypeBind && strings.HasPrefix(m.Source, cacheDir+"/")
}

func isArtifactMount(m mount.Mount) bool {
	return IsArtifactMount(&m)
}

// Attach gives a service spec exactly these artifact mounts, in place of any it
// had - an older version's, or another app's that a clone copied.
func Attach(spec *swarm.ServiceSpec, mounts []mount.Mount) {
	container := spec.TaskTemplate.ContainerSpec
	if container == nil {
		container = &swarm.ContainerSpec{}
		spec.TaskTemplate.ContainerSpec = container
	}
	container.Mounts = append(slices.DeleteFunc(container.Mounts, isArtifactMount), mounts...)
}

// KeepArtifactMounts is what a screen that rewrites an app's mounts saves: the
// mounts it built, and the artifacts the service already has. The storage screen
// does not show artifacts, so what it sends back never has them - and one it
// does send is not its to add.
func KeepArtifactMounts(final, current []mount.Mount) []mount.Mount {
	kept := slices.DeleteFunc(slices.Clone(final), isArtifactMount)
	for _, m := range current {
		if isArtifactMount(m) {
			kept = append(kept, m)
		}
	}
	return kept
}

// FromTemplate is the setting an app created from a template keeps its
// artifacts in, and nil for a template that has none.
func FromTemplate(artifacts []*templatemodel.Artifact) *entity.AppArtifactsSettings {
	if len(artifacts) == 0 {
		return nil
	}
	settings := &entity.AppArtifactsSettings{}
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		files := make(map[base.CPUArch]*entity.AppArtifactFile, len(artifact.Files))
		for _, arch := range slices.Sorted(maps.Keys(artifact.Files)) {
			if file := artifact.Files[arch]; file != nil {
				files[arch] = &entity.AppArtifactFile{URL: file.URL, SHA256: file.SHA256, Size: file.Size}
			}
		}
		settings.Artifacts = append(settings.Artifacts, &entity.AppArtifact{
			Name: artifact.Name, Version: artifact.Version, MountPath: artifact.MountPath, Files: files,
		})
	}
	return settings
}
