// Package imagescanrun runs a scan's container on the node it runs on: Trivy,
// with the volume its DBs and its layer cache are on, reading the image from
// this node's Docker when it has the image, from its registry otherwise. The
// app runs it for its own node; an agent, through ImageScanRun, for its.
package imagescanrun

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// DataDir is where the volume is mounted in the container.
	DataDir = "/data"

	// OutputMax is how large a report may be. Findings only, one is a few KB;
	// thousands of findings stay well under it.
	OutputMax = 32 << 20
	// LogMax is how much of Trivy's log is kept: its end, where an error is.
	LogMax = 256 << 10

	// authDir is where the registry's credentials are copied in the container,
	// for DOCKER_CONFIG: a file, never an argument or the environment, which
	// the container's inspect shows.
	authDir = "/tmp/hivepaas-auth"

	dockerSocket = "/var/run/docker.sock"

	// authDirMode and authFileMode keep the credentials to the container's root.
	authDirMode  = 0o700
	authFileMode = 0o600

	// removeTimeout is how long removing the container may take once the scan
	// is over, whatever became of the caller's context.
	removeTimeout = 30 * time.Second
)

var (
	// outputEndTimeout is how long the container's streams may go on once it
	// has stopped; a var, for the tests.
	outputEndTimeout = removeTimeout

	// imageIDPattern is an image's ID as Docker gives it.
	imageIDPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	// platformPattern is os/arch, with a variant or not: linux/arm64/v8.
	platformPattern = regexp.MustCompile(`^[a-z0-9]+/[a-z0-9_]+(/[a-z0-9]+)?$`)
)

// FormatCycloneDX asks for the image's SBOM, in place of its vulnerabilities.
const FormatCycloneDX = "cyclonedx"

// ModeRootfs scans a file of the artifacts cache in place of an image, as a
// root filesystem of one file: Trivy reads a Go binary's modules there. The
// file is the one the cache keeps by ImageID, its sha256.
const ModeRootfs = "rootfs"

// artifactTarget is where an artifact's file is bound in the container.
const artifactTarget = "/scan/artifact"

// RunReq is a scan of one image.
type RunReq struct {
	// Format is empty for a scan's report, FormatCycloneDX for an SBOM.
	Format string
	// Mode is empty for an image, ModeRootfs for an artifact's file.
	Mode string
	// TrivyImage is what the container runs; pulled when the node lacks it.
	TrivyImage string
	// ImageID is the image's ID as the node running it gives it.
	ImageID string
	// ImageNode names that node, for the error when neither this node nor a
	// registry has the image.
	ImageNode string
	// RemoteRef is the image's repository and digest, read when this node does
	// not have ImageID; empty when the image is in no registry.
	RemoteRef string
	// Platform is that node's, os/arch[/variant]: what is read of an image of
	// several platforms. Without it Trivy reads linux/amd64 on any node.
	Platform string
	// DockerConfig is a Docker config.json with the credentials of RemoteRef's
	// registry; nil for none.
	DockerConfig []byte
	// Volume is the mount of the volume Trivy's data is on; its target is
	// DataDir. CacheDir is Trivy's cache directory, below DataDir.
	Volume   mount.Mount
	CacheDir string
	// DBRepository and JavaDBRepository are mirrors of Trivy's DBs; empty for
	// Trivy's own.
	DBRepository     string
	JavaDBRepository string
	NanoCPUs         int64
	MemoryBytes      int64
	// Timeout is Trivy's.
	Timeout time.Duration
}

// RunResp is what the scan's container did.
type RunResp struct {
	// Source is where the image was read.
	Source base.ImageScanSource
	// Output is what the container printed: an imagescanservice.ScanOutput.
	Output []byte
	// Log is Trivy's log, its standard error; only its end when LogTruncated.
	Log          string
	LogTruncated bool
	// OOMKilled says the container ran out of its memory limit.
	OOMKilled bool
	ExitCode  int64
}

// Runner runs scans on the Docker it is given: the node's.
type Runner struct {
	Docker docker.Manager
}

func (r *Runner) Run(ctx context.Context, req *RunReq) (*RunResp, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	source, err := r.source(ctx, req)
	if err != nil {
		return nil, err
	}
	insecure := false
	if source == base.ImageScanSourceRegistry {
		insecure = r.insecureRegistry(ctx, req.RemoteRef)
	}
	if err = r.ensureImage(ctx, req.TrivyImage); err != nil {
		return nil, err
	}

	created, err := r.Docker.ContainerCreate(ctx, func(opts *client.ContainerCreateOptions) {
		*opts = containerOptions(req, source, insecure)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	containerID := created.ID
	defer func() {
		removeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), removeTimeout)
		defer cancel()
		_, _ = r.Docker.ContainerRemove(removeCtx, containerID, func(opts *client.ContainerRemoveOptions) {
			opts.Force = true
		})
	}()

	if source == base.ImageScanSourceRegistry && len(req.DockerConfig) > 0 {
		if err = r.copyCredentials(ctx, containerID, req.DockerConfig); err != nil {
			return nil, err
		}
	}
	return r.runContainer(ctx, containerID, req, source)
}

// runContainer starts the container, attached first so that all it prints is
// read though its logging driver keeps nothing, and waits for it.
func (r *Runner) runContainer(ctx context.Context, containerID string, req *RunReq,
	source base.ImageScanSource) (*RunResp, error) {
	attached, err := r.Docker.ContainerAttach(ctx, containerID, func(opts *client.ContainerAttachOptions) {
		opts.Stream, opts.Stdout, opts.Stderr = true, true, true
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer attached.Close()

	stdout := &capBuffer{max: OutputMax}
	stderr := &tailBuffer{max: LogMax}
	copied := make(chan error, 1)
	go func() {
		var copyErr error
		defer func() { copied <- copyErr }()
		defer safego.RecoverTo(&copyErr)
		_, copyErr = stdcopy.StdCopy(stdout, stderr, attached.Reader)
	}()

	if _, err = r.Docker.ContainerStart(ctx, containerID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	// Trivy stops itself at its timeout; the minute more is the DBs' check and
	// the container's own start and stop.
	waitCtx, cancel := context.WithTimeout(ctx, req.Timeout+time.Minute)
	defer cancel()
	waited := r.Docker.ContainerWait(waitCtx, containerID, func(opts *client.ContainerWaitOptions) {
		opts.Condition = container.WaitConditionNotRunning
	})
	resp := &RunResp{Source: source}
	select {
	case result := <-waited.Result:
		resp.ExitCode = result.StatusCode
	case err = <-waited.Error:
		return nil, hperrors.Wrap(err)
	case <-waitCtx.Done():
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("The scan did not end within %v.", req.Timeout+time.Minute)
	}
	// The streams end with the container. Ones that have not ended by then are
	// cut, and their copy waited for: the buffers are read only once nothing
	// writes them, and a report cut short is no report.
	select {
	case err = <-copied:
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	case <-time.After(outputEndTimeout):
		attached.Close()
		<-copied
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("The scan's output did not end within %v of the scan.", outputEndTimeout)
	}
	if stdout.over {
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("The report is larger than %v.", unit.DataSize(OutputMax))
	}
	if inspect, inspectErr := r.Docker.ContainerInspect(ctx, containerID); inspectErr == nil &&
		inspect.Container.State != nil {
		resp.OOMKilled = inspect.Container.State.OOMKilled
	}
	resp.Output = stdout.Bytes()
	resp.Log, resp.LogTruncated = stderr.String(), stderr.dropped
	return resp, nil
}

// source is where the image is read: this node's Docker when it has the
// image, its registry otherwise.
func (r *Runner) source(ctx context.Context, req *RunReq) (base.ImageScanSource, error) {
	if req.Mode == ModeRootfs {
		return base.ImageScanSourceArtifact, nil
	}
	_, err := r.Docker.ImageInspect(ctx, req.ImageID)
	switch {
	case err == nil:
		return base.ImageScanSourceNode, nil
	case !errors.Is(err, hperrors.ErrInfraNotFound):
		return "", hperrors.Wrap(err)
	case req.RemoteRef != "":
		return base.ImageScanSourceRegistry, nil
	}
	return "", hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail(
		"The image is on node %s only, and in no registry: the node scans run on cannot read it.",
		gofn.Coalesce(req.ImageNode, "?"))
}

// insecureRegistry says this node's Docker is told the registry speaks plain
// HTTP, as Trivy must be.
func (r *Runner) insecureRegistry(ctx context.Context, ref string) bool {
	named, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return false
	}
	info, err := r.Docker.SystemInfo(ctx)
	if err != nil || info.Info.RegistryConfig == nil {
		return false
	}
	index := info.Info.RegistryConfig.IndexConfigs[reference.Domain(named)]
	return index != nil && !index.Secure
}

func (r *Runner) ensureImage(ctx context.Context, image string) error {
	_, err := r.Docker.ImageInspect(ctx, image)
	if err == nil {
		return nil
	}
	if !errors.Is(err, hperrors.ErrInfraNotFound) {
		return hperrors.Wrap(err)
	}
	pull, err := r.Docker.ImagePull(ctx, image)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = pull.Close() }()
	return hperrors.Wrap(pull.Wait(ctx))
}

// copyCredentials puts config.json in the container before it starts, in a
// directory of its own the tar makes.
func (r *Runner) copyCredentials(ctx context.Context, containerID string, config []byte) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	dir := "hivepaas-auth"
	err := errors.Join(
		tw.WriteHeader(&tar.Header{Typeflag: tar.TypeDir, Name: dir + "/", Mode: authDirMode}),
		tw.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: dir + "/config.json", Mode: authFileMode,
			Size: int64(len(config))}),
	)
	if err == nil {
		_, err = tw.Write(config)
	}
	if err = errors.Join(err, tw.Close()); err != nil {
		return hperrors.Wrap(err)
	}
	_, err = r.Docker.ContainerCopyTo(ctx, containerID, "/tmp", &buf)
	return hperrors.Wrap(err)
}

// containerOptions is the scan's container: Trivy's image under the scan's
// limits, the volume, the node's Docker socket only to read an image from it,
// and no log the node keeps - the collector of stored logs reads every
// container's. Its name and labels make it a temporary one, which the cluster
// cleanup removes should the scan not.
func containerOptions(req *RunReq, source base.ImageScanSource, insecure bool) client.ContainerCreateOptions {
	volume := req.Volume
	volume.Target = DataDir
	mounts := []mount.Mount{volume}
	var env []string
	switch {
	case source == base.ImageScanSourceArtifact:
		mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: artifactPath(req.ImageID),
			Target: artifactTarget, ReadOnly: true})
	case source == base.ImageScanSourceNode:
		mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: dockerSocket, Target: dockerSocket})
	case len(req.DockerConfig) > 0:
		env = append(env, "DOCKER_CONFIG="+authDir)
	}
	return client.ContainerCreateOptions{
		Name: docker.TempContainerPrefix + "scan-" + gofn.RandString(8), //nolint:mnd
		Config: &container.Config{
			Image:      req.TrivyImage,
			Entrypoint: []string{"sh", "-c"},
			Cmd:        []string{Script(req, source, insecure)},
			Env:        env,
			Labels: map[string]string{
				docker.LabelTempResource:  docker.LabelTempResourceVal,
				docker.LabelTempCreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
		HostConfig: &container.HostConfig{
			Mounts:    mounts,
			LogConfig: container.LogConfig{Type: "none"},
			Resources: container.Resources{NanoCPUs: req.NanoCPUs, Memory: req.MemoryBytes},
		},
	}
}

func (req *RunReq) validate() error {
	invalid := func(format string, args ...any) error {
		return hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail(format, args...)
	}
	if !imageIDPattern.MatchString(req.ImageID) {
		return invalid("%q is not an image's ID.", req.ImageID)
	}
	if req.RemoteRef != "" {
		named, err := reference.ParseNormalizedNamed(req.RemoteRef)
		if err != nil {
			return invalid("%q is not a repository and digest.", req.RemoteRef)
		}
		if _, canonical := named.(reference.Canonical); !canonical {
			return invalid("%q is not a repository and digest.", req.RemoteRef)
		}
		if !platformPattern.MatchString(req.Platform) {
			return invalid("%q is not a platform.", req.Platform)
		}
	}
	if req.Format != "" && req.Format != FormatCycloneDX {
		return invalid("%q is no format a scan writes.", req.Format)
	}
	switch req.Mode {
	case "":
	case ModeRootfs:
		if req.RemoteRef != "" {
			return invalid("An artifact is read from the artifacts cache, not from a registry.")
		}
		if artifactservice.CacheHostDir() == "" {
			return invalid("This node keeps no artifacts: no host directory is configured.")
		}
	default:
		return invalid("%q is no mode a scan runs in.", req.Mode)
	}
	if req.TrivyImage == "" || req.CacheDir == "" || req.Volume.Source == "" || req.Timeout <= 0 {
		return invalid("The scan is missing its image, its volume or its timeout.")
	}
	return nil
}

// artifactPath is the file of the artifacts cache an artifact's ID names: the
// cache keeps each by its sha256, which the ID is.
func artifactPath(id string) string {
	return filepath.Join(artifactservice.CacheHostDir(), strings.TrimPrefix(id, "sha256:"))
}

// capBuffer keeps what is written up to max, and says when there was more.
type capBuffer struct {
	bytes.Buffer
	max  int
	over bool
}

func (b *capBuffer) Write(p []byte) (int, error) {
	if room := b.max - b.Len(); len(p) > room {
		b.over = true
		_, _ = b.Buffer.Write(p[:max(room, 0)])
		return len(p), nil
	}
	return b.Buffer.Write(p) //nolint:wrapcheck // a bytes.Buffer
}

// tailBuffer keeps the last max bytes written.
type tailBuffer struct {
	bytes.Buffer
	max     int
	dropped bool
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	n, _ := b.Buffer.Write(p)
	if extra := b.Len() - b.max; extra > 0 {
		b.Next(extra)
		b.dropped = true
	}
	return n, nil
}
