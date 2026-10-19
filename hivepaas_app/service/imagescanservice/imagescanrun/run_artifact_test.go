package imagescanrun

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const artifactSHA = "965a7147852435bc4b54f272799b63df826b54ef65cf7c4fc688c82b96c0c4eb"

func useHostDir(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Storage: config.Storage{HostDir: "/var/lib/hivepaas"}})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func artifactReq() *RunReq {
	req := scanReq()
	req.Mode, req.ImageID, req.ImageNode = ModeRootfs, "sha256:"+artifactSHA, "n1"
	return req
}

// An artifact is scanned as a root filesystem of one file, at a path of its
// own; its SBOM likewise.
func TestScriptOfAnArtifact(t *testing.T) {
	req := artifactReq()
	req.DBRepository = "mirror.local:5000/aquasec/trivy-db:2"
	script := Script(req, base.ImageScanSourceArtifact, false)
	for _, part := range []string{
		"trivy rootfs --cache-dir /data/hivepaas/scanning/trivy", "--scanners vuln --list-all-pkgs=false",
		"--format json", "--db-repository mirror.local:5000/aquasec/trivy-db:2",
		artifactTarget + " > /tmp/report.json",
	} {
		assert.Contains(t, script, part)
	}
	assert.NotContains(t, script, "--image-src")
	assert.NotContains(t, script, "trivy image")

	req.Format = FormatCycloneDX
	script = Script(req, base.ImageScanSourceArtifact, false)
	assert.Contains(t, script, "trivy rootfs")
	assert.Contains(t, script, "--format cyclonedx")
	assert.NotContains(t, script, "--scanners")
}

// An artifact is a file of the artifacts cache by its hash, never a path a
// call names, and is in no registry.
func TestAnArtifactsRunReqIsValidated(t *testing.T) {
	useHostDir(t)
	for name, change := range map[string]func(*RunReq){
		"another mode":  func(r *RunReq) { r.Mode = "fs" },
		"not its hash":  func(r *RunReq) { r.ImageID = "sha256:../../etc/passwd" },
		"in a registry": func(r *RunReq) { r.RemoteRef, r.Platform = "alpine@"+indexDigest, "linux/amd64" },
	} {
		req := artifactReq()
		change(req)
		assert.ErrorIs(t, req.validate(), hperrors.ErrImageScanFailed, name)
	}
	assert.NoError(t, artifactReq().validate())

	config.SetCurrent(&config.Config{})
	assert.ErrorIs(t, artifactReq().validate(), hperrors.ErrImageScanFailed, "no artifacts cache")
}

// The file is bound read-only from the node's artifacts cache, and nothing of
// the node's Docker is asked or given: no socket, no image looked up.
func TestRunScansAnArtifactFromTheCache(t *testing.T) {
	useHostDir(t)
	n := &node{images: map[string]bool{trivy: true}, stdout: `{"exitCode":0,"report":{}}`}

	resp, err := (&Runner{Docker: n}).Run(context.Background(), artifactReq())

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, base.ImageScanSourceArtifact, resp.Source)
	assert.Equal(t, []mount.Mount{
		{Type: mount.TypeVolume, Source: "trivy-data", Target: DataDir},
		{Type: mount.TypeBind, Source: "/var/lib/hivepaas/cache/artifacts/sha256/" + artifactSHA,
			Target: artifactTarget, ReadOnly: true},
	}, n.created.HostConfig.Mounts)
	assert.Contains(t, n.created.Config.Cmd[0], "trivy rootfs")
}
