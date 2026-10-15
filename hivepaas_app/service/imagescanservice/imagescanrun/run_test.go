package imagescanrun

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"iter"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	trivy       = "aquasec/trivy:0.75.0"
	imageID     = "sha256:19a85ffa0ee14b7c64ca5104dc4e1c5021fa04bd6b4a692074691ce154f5d074"
	indexDigest = "sha256:451eee8bedcb2f029756dc3e9d73bab0e7943c1ac55cff3a4861c52a0fdd3e98"
)

func scanReq() *RunReq {
	return &RunReq{
		TrivyImage: trivy,
		ImageID:    imageID,
		ImageNode:  "n2",
		Volume:     mount.Mount{Type: mount.TypeVolume, Source: "trivy-data"},
		CacheDir:   "/data/hivepaas/scanning/trivy",
		Timeout:    10 * time.Minute,
	}
}

// The script scans the image from the node's Docker by its ID, or from its
// registry by digest for the platform of the node running it; and prints the
// exit code, trivy version and the report as one object.
func TestScript(t *testing.T) {
	req := scanReq()
	req.JavaDBRepository = "mirror.local:5000/aquasec/trivy-java-db:1"
	script := Script(req, base.ImageScanSourceNode, false)
	for _, part := range []string{
		`find /data/hivepaas/scanning/trivy/java-db/metadata.json -mtime -7`, `skip=--skip-java-db-update`,
		"trivy image --cache-dir /data/hivepaas/scanning/trivy", "--scanners vuln --list-all-pkgs=false",
		"--disable-telemetry", "--skip-version-check", "--no-progress", "--format json", "--timeout 10m0s",
		"$skip", "--java-db-repository mirror.local:5000/aquasec/trivy-java-db:1",
		"--image-src docker " + imageID + " > /tmp/report.json",
		"trivy version --format json --cache-dir /data/hivepaas/scanning/trivy",
		`printf '{"exitCode":%d,"version":%s,"report":' "$code" "$version"`,
		`exit "$code"`,
	} {
		assert.Contains(t, script, part)
	}
	assert.NotContains(t, script, "--platform")
	assert.NotContains(t, script, "--db-repository ")

	req.Format = FormatCycloneDX
	script = Script(req, base.ImageScanSourceNode, false)
	assert.Contains(t, script, "--format cyclonedx")
	assert.NotContains(t, script, "--scanners", "an SBOM needs no DB")
	assert.NotContains(t, script, "--java-db-repository")
	req.Format = ""

	req.RemoteRef, req.Platform = "registry.local:5000/shop/api@"+indexDigest, "linux/arm64/v8"
	script = Script(req, base.ImageScanSourceRegistry, true)
	assert.Contains(t, script, "--image-src remote --platform linux/arm64/v8 --insecure registry.local:5000/shop/api@"+
		indexDigest+" > /tmp/report.json")
}

// Only an image's ID, a repository with a digest and a platform reach the
// command line.
func TestRunReqIsValidated(t *testing.T) {
	for _, change := range []func(*RunReq){
		func(r *RunReq) { r.ImageID = "alpine:3.10; rm -rf /" },
		func(r *RunReq) { r.RemoteRef, r.Platform = "alpine:3.10", "linux/amd64" },
		func(r *RunReq) { r.RemoteRef, r.Platform = "alpine@"+indexDigest, "linux/amd64; id" },
		func(r *RunReq) { r.Volume = mount.Mount{} },
		func(r *RunReq) { r.Format = "spdx; id" },
	} {
		req := scanReq()
		change(req)
		assert.ErrorIs(t, req.validate(), hperrors.ErrImageScanFailed)
	}
	assert.NoError(t, scanReq().validate())
}

// node is a node's Docker, which has some images and runs a container that
// prints what it is given.
type node struct {
	docker.Manager
	images    map[string]bool
	insecure  string
	stdout    string
	stderr    string
	exitCode  int64
	oomKilled bool

	created *client.ContainerCreateOptions
	copied  map[string]string
	removed bool
	pulled  []string
}

func (n *node) ImageInspect(_ context.Context, image string, _ ...docker.ImageInspectOption) (
	*client.ImageInspectResult, error) {
	if !n.images[image] {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	return &client.ImageInspectResult{}, nil
}

func (n *node) ImagePull(_ context.Context, image string, _ ...docker.ImagePullOption) (
	client.ImagePullResponse, error) {
	n.pulled = append(n.pulled, image)
	return pulled{io.NopCloser(strings.NewReader(""))}, nil
}

func (n *node) SystemInfo(context.Context, ...docker.SystemInfoOption) (*client.SystemInfoResult, error) {
	info := &client.SystemInfoResult{Info: system.Info{RegistryConfig: &registry.ServiceConfig{
		IndexConfigs: map[string]*registry.IndexInfo{}}}}
	if n.insecure != "" {
		info.Info.RegistryConfig.IndexConfigs[n.insecure] = &registry.IndexInfo{Name: n.insecure, Secure: false}
	}
	return info, nil
}

func (n *node) ContainerCreate(_ context.Context, options ...docker.ContainerCreateOption) (
	*client.ContainerCreateResult, error) {
	opts := client.ContainerCreateOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	n.created = &opts
	return &client.ContainerCreateResult{ID: "c1"}, nil
}

func (n *node) ContainerCopyTo(_ context.Context, _ string, dst string, content io.Reader,
	_ ...docker.ContainerCopyToOption) (*client.CopyToContainerResult, error) {
	n.copied = map[string]string{}
	tr := tar.NewReader(content)
	for header, err := tr.Next(); err == nil; header, err = tr.Next() {
		data, _ := io.ReadAll(tr)
		n.copied[dst+"/"+header.Name] = string(data)
	}
	return &client.CopyToContainerResult{}, nil
}

func (n *node) ContainerAttach(context.Context, string, ...docker.ContainerAttachOption) (
	*client.ContainerAttachResult, error) {
	var stream bytes.Buffer
	writeFrame(&stream, stdcopy.Stdout, n.stdout)
	writeFrame(&stream, stdcopy.Stderr, n.stderr)
	conn, _ := net.Pipe()
	resp := &client.ContainerAttachResult{}
	resp.Conn, resp.Reader = conn, bufio.NewReader(&stream)
	return resp, nil
}

func (n *node) ContainerStart(context.Context, string, ...docker.ContainerStartOption) (
	*client.ContainerStartResult, error) {
	return &client.ContainerStartResult{}, nil
}

func (n *node) ContainerWait(context.Context, string, ...docker.ContainerWaitOption) *client.ContainerWaitResult {
	result := make(chan container.WaitResponse, 1)
	result <- container.WaitResponse{StatusCode: n.exitCode}
	return &client.ContainerWaitResult{Result: result, Error: make(chan error)}
}

func (n *node) ContainerInspect(context.Context, string, ...docker.ContainerInspectOption) (
	*client.ContainerInspectResult, error) {
	return &client.ContainerInspectResult{Container: container.InspectResponse{
		State: &container.State{OOMKilled: n.oomKilled}}}, nil
}

func (n *node) ContainerRemove(context.Context, string, ...docker.ContainerRemoveOption) (
	*client.ContainerRemoveResult, error) {
	n.removed = true
	return &client.ContainerRemoveResult{}, nil
}

// writeFrame writes data as Docker multiplexes a container's streams: a
// header of the stream and the length, then the bytes.
func writeFrame(w *bytes.Buffer, stream stdcopy.StdType, data string) {
	header := make([]byte, 8) //nolint:mnd // the header's size
	header[0] = byte(stream)
	binary.BigEndian.PutUint32(header[4:], uint32(len(data))) //nolint:gosec // a test's few bytes
	w.Write(header)
	w.WriteString(data)
}

// pulled is a pull that has ended.
type pulled struct{ io.ReadCloser }

func (pulled) JSONMessages(context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(func(jsonstream.Message, error) bool) {}
}

func (pulled) Wait(context.Context) error { return nil }

// A node that has the image reads it from its own Docker: the container gets
// the socket and the volume, its log stays out of the node's, and what it
// prints comes back.
func TestRunReadsTheNodesImage(t *testing.T) {
	n := &node{images: map[string]bool{imageID: true, trivy: true},
		stdout: `{"exitCode":0,"report":{}}`, stderr: "INFO scanning\n"}
	req := scanReq()
	req.NanoCPUs, req.MemoryBytes = 1e9, 512<<20
	resp, err := (&Runner{Docker: n}).Run(context.Background(), req)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, base.ImageScanSourceNode, resp.Source)
	assert.Equal(t, `{"exitCode":0,"report":{}}`, string(resp.Output))
	assert.Equal(t, "INFO scanning\n", resp.Log)
	assert.True(t, n.removed)
	assert.Empty(t, n.pulled, "the node has Trivy's image")

	created := n.created
	assert.Equal(t, []string{"sh", "-c"}, created.Config.Entrypoint)
	assert.True(t, strings.HasPrefix(created.Name, docker.TempContainerPrefix+"scan-"))
	assert.Equal(t, docker.LabelTempResourceVal, created.Config.Labels[docker.LabelTempResource])
	assert.Equal(t, "none", created.HostConfig.LogConfig.Type)
	assert.Equal(t, int64(512<<20), created.HostConfig.Memory)
	assert.Equal(t, []mount.Mount{
		{Type: mount.TypeVolume, Source: "trivy-data", Target: DataDir},
		{Type: mount.TypeBind, Source: dockerSocket, Target: dockerSocket},
	}, created.HostConfig.Mounts)
	assert.Empty(t, created.Config.Env)
}

// A node without the image reads it from its registry, with the credentials
// copied in, never in its environment's values or its command; without the
// socket. A registry the node's Docker calls insecure is read over HTTP.
func TestRunReadsTheImageFromItsRegistry(t *testing.T) {
	n := &node{images: map[string]bool{}, insecure: "registry.local:5000", stdout: `{"exitCode":0}`}
	req := scanReq()
	req.RemoteRef, req.Platform = "registry.local:5000/shop/api@"+indexDigest, "linux/arm64"
	req.DockerConfig = []byte(`{"auths":{"registry.local:5000":{"auth":"dTpw"}}}`)
	resp, err := (&Runner{Docker: n}).Run(context.Background(), req)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, base.ImageScanSourceRegistry, resp.Source)
	assert.Equal(t, []string{trivy}, n.pulled, "Trivy's image, pulled where the node lacks it")
	assert.Equal(t, string(req.DockerConfig), n.copied["/tmp/hivepaas-auth/config.json"])
	assert.Equal(t, []string{"DOCKER_CONFIG=" + authDir}, n.created.Config.Env)
	assert.Len(t, n.created.HostConfig.Mounts, 1, "no socket")
	assert.Contains(t, n.created.Config.Cmd[0], "--insecure registry.local:5000/shop/api@")
	assert.NotContains(t, n.created.Config.Cmd[0], "dTpw")
}

// An image on another node only, and in no registry, cannot be read here.
func TestRunRefusesAnImageNoRegistryHas(t *testing.T) {
	_, err := (&Runner{Docker: &node{images: map[string]bool{}}}).Run(context.Background(), scanReq())
	assert.ErrorIs(t, err, hperrors.ErrImageScanFailed)
	assert.Contains(t, hperrors.GetErrorDetail(err, ""), "on node n2 only, and in no registry")
}

// A report larger than the limit fails the scan; the log keeps its end.
func TestBuffers(t *testing.T) {
	out := &capBuffer{max: 4}
	_, _ = out.Write([]byte("abc"))
	_, _ = out.Write([]byte("def"))
	assert.True(t, out.over)
	assert.Equal(t, "abcd", out.String())

	log := &tailBuffer{max: 4}
	_, _ = log.Write([]byte("abc"))
	_, _ = log.Write([]byte("def"))
	assert.True(t, log.dropped)
	assert.Equal(t, "cdef", log.String())
}
