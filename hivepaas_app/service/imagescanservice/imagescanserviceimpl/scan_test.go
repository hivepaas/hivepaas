package imagescanserviceimpl

import (
	"strings"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

const imageID = "sha256:19a85ffa0ee14b7c64ca5104dc4e1c5021fa04bd6b4a692074691ce154f5d074"

// A scan asks the server with the token's file, reads the node's Docker, keeps
// its report to findings and renews the Java DB only once the node's is a week
// old.
func TestScanScript(t *testing.T) {
	script, err := scanScript(&entity.ImageScanSettings{JavaDBRepository: "mirror.local:5000/aquasec/trivy-java-db:1"},
		imageID)
	assert.NoError(t, err)
	assert.Contains(t, script, `find /root/.cache/trivy/java-db/metadata.json -mtime -7`)
	assert.Contains(t, script, `skip=--skip-java-db-update`)
	for _, arg := range []string{"exec trivy image", "--server http://trivy:4954",
		`--token "$(cat /run/hivepaas/trivy-token)"`, "--image-src docker", "--scanners vuln",
		"--list-all-pkgs=false", "--disable-telemetry", "--skip-version-check", "--format json",
		"--timeout 10m0s", "$skip", "--java-db-repository mirror.local:5000/aquasec/trivy-java-db:1"} {
		assert.Contains(t, script, arg)
	}
	assert.True(t, strings.HasSuffix(script, " "+imageID))

	script, err = scanScript(&entity.ImageScanSettings{}, imageID)
	assert.NoError(t, err)
	assert.NotContains(t, script, "--java-db-repository")

	_, err = scanScript(&entity.ImageScanSettings{}, "alpine:3.10; rm -rf /")
	assert.Error(t, err, "only an image's ID reaches the command line")
}

// docker inspect's lines are an image's ID and the reference the container was
// made with; anything else is passed over.
func TestParseContainerImages(t *testing.T) {
	out := imageID + " registry.local/shop/api:abc123@sha256:aaaa\n" +
		"Error: No such object: c9\n" +
		imageID + " registry.local/shop/api:abc123\n"
	images := parseContainerImages(out, "n1")
	assert.Equal(t, []runningImage{
		{ID: imageID, Ref: "registry.local/shop/api:abc123", NodeID: "n1"},
		{ID: imageID, Ref: "registry.local/shop/api:abc123", NodeID: "n1"},
	}, images)
}

// Trivy's error is shown without its time and level.
func TestTrivyMessage(t *testing.T) {
	assert.Equal(t, "Fatal error	run error: twirp error unauthenticated: invalid token",
		trivyMessage("2026-10-10T18:18:09Z\tFATAL\tFatal error\trun error: twirp error unauthenticated: invalid token"))
	assert.Equal(t, "plain", trivyMessage(" plain "))
}

// The server reads its token from the file, and its DB from a mirror when
// told: the command survives the shell split a deployment does.
func TestServerCommand(t *testing.T) {
	assert.Equal(t, `-c 'exec trivy server --listen 0.0.0.0:4954 --cache-dir /data --token "$(cat `+
		`/run/hivepaas/trivy-token)"'`, serverCommand(""))
	assert.Contains(t, serverCommand("mirror.local:5000/aquasec/trivy-db:2"),
		"--db-repository mirror.local:5000/aquasec/trivy-db:2")
}

// Both apps' documents are buildable as a template's are, with the token as a
// file and the limits given.
func TestBuildAppDoc(t *testing.T) {
	doc, err := buildAppDoc(appDocInput{Image: "aquasec/trivy:0.75.0", Entrypoint: "sh", Command: runnerCommand(),
		Token: "t0k", Resources: systemappservice.Resources{CPULimit: 1, MemoryLimit: (512 * unit.MB).Bytes()}})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "sh", doc.Deployment.Source["entrypoint"])
	assert.Equal(t, 1.0, doc.Deployment.Resources.Limits.CPUs)
	assert.Equal(t, 512*unit.MB, doc.Deployment.Resources.Limits.Memory)
	secrets := doc.Settings["secrets"].(map[string]any)
	assert.Equal(t, "t0k", secrets[tokenKey].(map[string]any)["value"])
}

// The runner is one task a node, with the Java DB's volume; the server keeps
// its DB in one. Customizing twice adds a volume once.
func TestCustomize(t *testing.T) {
	spec := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
		Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{}}}
	assert.NoError(t, customizeRunner(spec))
	assert.NoError(t, customizeRunner(spec))
	assert.NotNil(t, spec.Mode.Global)
	assert.Nil(t, spec.Mode.Replicated)
	assert.Equal(t, []mount.Mount{{Type: mount.TypeVolume, Source: runnerCacheVolume, Target: runnerCacheDir}},
		spec.TaskTemplate.ContainerSpec.Mounts)

	server := &swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}}}
	assert.NoError(t, customizeServer(server))
	assert.Equal(t, serverDataVolume, server.TaskTemplate.ContainerSpec.Mounts[0].Source)
}
