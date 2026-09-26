package specmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	// MaxDockerAPIImages is how many image patterns a document may give an app's
	// children. A job runner takes "*"; an app that runs a few of its own names
	// them.
	MaxDockerAPIImages = 20
	// MaxDockerAPISharedDirs is how many of its directories an app may share.
	MaxDockerAPISharedDirs = 5
	// MaxDockerAPISharedVolumes is how many volume names may stand for them.
	MaxDockerAPISharedVolumes = 5
	// MaxDockerAPIContainers is the most children one app may keep at once.
	MaxDockerAPIContainers = 100
	// minDockerAPIMemory is the least memory docker starts a container with.
	minDockerAPIMemory = 6 * unit.MB

	dockerAPIPrefix = "settings.dockerApi."
)

var (
	// volumeNamePattern is Docker's own rule for a volume's name.
	volumeNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,127}$`)

	dockerAPIModes    = []string{"", entity.DockerAPIModeProxy, entity.DockerAPIModeHost}
	dockerAPINetworks = []string{entity.DockerAPINetworkEnv}
	dockerAPIGroups   = []string{string(dockerproxy.GroupExec), string(dockerproxy.GroupFiles),
		string(dockerproxy.GroupVolumes), string(dockerproxy.GroupNetworks), string(dockerproxy.GroupNestedSocket)}
)

// DockerAPIIn reads the dockerApi block of a document's settings, and is nil
// when there is none. A field the setting does not have is refused rather than
// dropped: it would be a grant that is not applied.
func DockerAPIIn(settings map[string]any) (*entity.AppDockerAPISettings, error) {
	body, found := settings[SingletonBlockName(base.SettingTypeAppDockerAPI)]
	if !found {
		return nil, nil
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, invalid("%s%s", dockerAPIPrefix, err.Error())
	}
	out := &entity.AppDockerAPISettings{}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(out); err != nil {
		return nil, invalid("%s%s", dockerAPIPrefix, err.Error())
	}
	return out, nil
}

// DockerAPIProblem says what is wrong with a Docker API block, and is empty when
// nothing is. In host mode the proxy's policy is kept but not used, and not
// checked either: going back to the proxy checks it.
func DockerAPIProblem(s *entity.AppDockerAPISettings) string {
	switch {
	case s == nil:
		return ""
	case !slices.Contains(dockerAPIModes, s.Mode):
		return fmt.Sprintf("%smode: %q is not one of %s or %s",
			dockerAPIPrefix, s.Mode, entity.DockerAPIModeProxy, entity.DockerAPIModeHost)
	case s.IsHostMode():
		return ""
	}
	if problem := dockerAPICountProblem(s); problem != "" {
		return problem
	}
	for i, image := range s.Images {
		if strings.TrimSpace(image) == "" || strings.ContainsAny(image, " \t\n") {
			return fmt.Sprintf("%simages[%d]: %q is not an image", dockerAPIPrefix, i, image)
		}
	}
	for i, dir := range s.SharedDirs {
		if !path.IsAbs(dir) || path.Clean(dir) != dir || dir == "/" {
			return fmt.Sprintf("%ssharedDirs[%d]: %s is not an absolute path to a directory below /",
				dockerAPIPrefix, i, dir)
		}
	}
	if problem := sharedVolumesProblem(s); problem != "" {
		return problem
	}
	for i, network := range s.Networks {
		if !slices.Contains(dockerAPINetworks, network) {
			return fmt.Sprintf("%snetworks[%d]: %q is not one of %v", dockerAPIPrefix, i, network, dockerAPINetworks)
		}
	}
	for i, group := range s.Allow {
		if !slices.Contains(dockerAPIGroups, group) {
			return fmt.Sprintf("%sallow[%d]: %q is not one of %v", dockerAPIPrefix, i, group, dockerAPIGroups)
		}
	}
	return dockerAPILimitsProblem(s.Limits)
}

func dockerAPICountProblem(s *entity.AppDockerAPISettings) string {
	switch {
	case len(s.Images) == 0:
		return dockerAPIPrefix + "images: at least one image is needed"
	case len(s.Images) > MaxDockerAPIImages:
		return fmt.Sprintf("%simages: at most %d, and this has %d", dockerAPIPrefix, MaxDockerAPIImages, len(s.Images))
	case len(s.SharedDirs) > MaxDockerAPISharedDirs:
		return fmt.Sprintf("%ssharedDirs: at most %d, and this has %d",
			dockerAPIPrefix, MaxDockerAPISharedDirs, len(s.SharedDirs))
	case len(s.SharedVolumes) > MaxDockerAPISharedVolumes:
		return fmt.Sprintf("%ssharedVolumes: at most %d, and this has %d",
			dockerAPIPrefix, MaxDockerAPISharedVolumes, len(s.SharedVolumes))
	}
	return ""
}

// sharedVolumesProblem refuses a shared volume that is not a volume's name, or
// that stands for a directory none of sharedDirs covers: a name gives what a
// bind of the path would, so it can give nothing sharedDirs does not.
//
// A name HivePaaS keeps for itself needs no rule of its own. The proxy reads a
// shared volume before anything else a name could be, and it gives the app's own
// directory, so one called like another app's socket volume only shadows it.
func sharedVolumesProblem(s *entity.AppDockerAPISettings) string {
	for _, name := range slices.Sorted(maps.Keys(s.SharedVolumes)) {
		dir := s.SharedVolumes[name]
		switch {
		case !volumeNamePattern.MatchString(name):
			return fmt.Sprintf("%ssharedVolumes: %q is not a volume name", dockerAPIPrefix, name)
		case !path.IsAbs(dir) || path.Clean(dir) != dir:
			return fmt.Sprintf("%ssharedVolumes.%s: %s is not an absolute path", dockerAPIPrefix, name, dir)
		case !slices.ContainsFunc(s.SharedDirs, func(shared string) bool { return coversPath(shared, dir) }):
			return fmt.Sprintf("%ssharedVolumes.%s: %s is in none of sharedDirs", dockerAPIPrefix, name, dir)
		}
	}
	return ""
}

// coversPath reports whether p is dir or lies below it.
func coversPath(dir, p string) bool {
	return p == dir || strings.HasPrefix(p, strings.TrimSuffix(dir, "/")+"/")
}

func dockerAPILimitsProblem(limits entity.AppDockerAPILimits) string {
	switch {
	case limits.Containers < 0 || limits.Containers > MaxDockerAPIContainers:
		return fmt.Sprintf("%slimits.containers: %d is not between 0 and %d",
			dockerAPIPrefix, limits.Containers, MaxDockerAPIContainers)
	case limits.Memory != 0 && limits.Memory < minDockerAPIMemory:
		return fmt.Sprintf("%slimits.memory: %s is less than the %s docker starts a container with",
			dockerAPIPrefix, limits.Memory, minDockerAPIMemory)
	case limits.CPUs < 0:
		return fmt.Sprintf("%slimits.cpus: %v is below zero", dockerAPIPrefix, limits.CPUs)
	}
	return ""
}

// HostModeFromTemplate is why a template may not give its app the node's own
// socket.
const HostModeFromTemplate = dockerAPIPrefix + "mode: host is given by an administrator, on the app's " +
	"Docker API screen, never by a template"

// checkDockerAPI refuses a block that is wrong in itself, and a shared
// directory the app does not keep on its own storage: what a child is given is
// the directory the app mounts there, so there has to be one.
func checkDockerAPI(doc *AppDoc) error {
	settings, err := DockerAPIIn(doc.Settings)
	if err != nil || settings == nil {
		return err
	}
	if problem := DockerAPIProblem(settings); problem != "" {
		return invalid("%s", problem)
	}
	if settings.IsHostMode() {
		return invalid("%s", HostModeFromTemplate)
	}
	var targets []string
	if doc.Deployment != nil && doc.Deployment.Storage != nil {
		targets = slices.Collect(maps.Keys(doc.Deployment.Storage.Mounts))
	}
	if problem := SharedDirsProblem(settings.SharedDirs, targets); problem != "" {
		return invalid("%s", problem)
	}
	return nil
}

// SharedDirsProblem names the first shared directory on none of the mount
// targets given, and is empty when each is on one.
func SharedDirsProblem(dirs, targets []string) string {
	for _, dir := range dirs {
		covered := slices.ContainsFunc(targets, func(target string) bool { return coversPath(target, dir) })
		if !covered {
			return fmt.Sprintf("%ssharedDirs: %s is on none of the app's storage mounts", dockerAPIPrefix, dir)
		}
	}
	return ""
}
