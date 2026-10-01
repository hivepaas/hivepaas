package functiontest

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"iter"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// copiedFile is a file a test run copied into its container.
type copiedFile struct {
	content string
	uid     int
	mode    int64
}

// fakeDocker is a node's docker: its images, and the one container a test run
// makes, which exits as the test says.
type fakeDocker struct {
	docker.Manager
	t *testing.T

	images map[string]bool
	pulled []string

	created *client.ContainerCreateOptions
	copied  map[string]copiedFile
	started bool
	killed  bool
	removed bool

	// exitCode is the container's, and never makes it run past any timeout.
	exitCode int64
	never    bool
	stdout   string
	stderr   string
	// files are what the container's /app holds once it exits, by name.
	files map[string]string
}

func newFakeDocker(t *testing.T) *fakeDocker {
	return &fakeDocker{t: t, images: map[string]bool{}, copied: map[string]copiedFile{}, files: map[string]string{}}
}

func (f *fakeDocker) ImageInspect(_ context.Context, image string, _ ...docker.ImageInspectOption) (
	*client.ImageInspectResult, error) {
	if f.images[image] {
		return &client.ImageInspectResult{}, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
}

type pullResponse struct{ io.ReadCloser }

func (pullResponse) Wait(context.Context) error { return nil }

func (pullResponse) JSONMessages(context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(func(jsonstream.Message, error) bool) {}
}

func (f *fakeDocker) ImagePull(_ context.Context, image string, _ ...docker.ImagePullOption) (
	client.ImagePullResponse, error) {
	f.pulled = append(f.pulled, image)
	f.images[image] = true
	return pullResponse{ReadCloser: io.NopCloser(&bytes.Buffer{})}, nil
}

func (f *fakeDocker) ContainerCreate(_ context.Context, options ...docker.ContainerCreateOption) (
	*client.ContainerCreateResult, error) {
	opts := &client.ContainerCreateOptions{}
	for _, opt := range options {
		opt(opts)
	}
	f.created = opts
	return &client.ContainerCreateResult{ID: "container-1"}, nil
}

func (f *fakeDocker) ContainerCopyTo(_ context.Context, _ string, dst string, content io.Reader,
	options ...docker.ContainerCopyToOption) (*client.CopyToContainerResult, error) {
	opts := &client.CopyToContainerOptions{}
	for _, opt := range options {
		opt(opts)
	}
	assert.True(f.t, opts.CopyUIDGID, "the files keep the owner the tar gives them")
	reader := tar.NewReader(content)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(f.t, err)
		body, _ := io.ReadAll(reader)
		f.copied[dst+"/"+header.Name] = copiedFile{content: string(body), uid: header.Uid, mode: header.Mode}
	}
	return &client.CopyToContainerResult{}, nil
}

func (f *fakeDocker) ContainerStart(_ context.Context, _ string, _ ...docker.ContainerStartOption) (
	*client.ContainerStartResult, error) {
	f.started = true
	return &client.ContainerStartResult{}, nil
}

func (f *fakeDocker) ContainerWait(
	_ context.Context, _ string, _ ...docker.ContainerWaitOption,
) *client.ContainerWaitResult {
	results := make(chan container.WaitResponse, 1)
	errs := make(chan error, 1)
	if !f.never {
		results <- container.WaitResponse{StatusCode: f.exitCode}
	}
	return &client.ContainerWaitResult{Result: results, Error: errs}
}

func (f *fakeDocker) ContainerKill(_ context.Context, _ string, _ string, _ ...docker.ContainerKillOption) (
	*client.ContainerKillResult, error) {
	f.killed = true
	return &client.ContainerKillResult{}, nil
}

func (f *fakeDocker) ContainerLogs(_ context.Context, _ string, _ ...docker.ContainerLogsOption) (
	client.ContainerLogsResult, error) {
	var buf bytes.Buffer
	frame(&buf, stdcopy.Stdout, f.stdout)
	frame(&buf, stdcopy.Stderr, f.stderr)
	return io.NopCloser(&buf), nil
}

// frame writes what a container wrote on one of its streams, as docker's logs
// carry it: an 8-byte header naming the stream and the size, then the bytes.
func frame(buf *bytes.Buffer, stream stdcopy.StdType, data string) {
	if data == "" {
		return
	}
	header := make([]byte, 8)
	header[0] = byte(stream)
	binary.BigEndian.PutUint32(header[4:], uint32(len(data)))
	buf.Write(header)
	buf.WriteString(data)
}

func (f *fakeDocker) ContainerCopyFrom(_ context.Context, _ string, path string) (
	*client.CopyFromContainerResult, error) {
	name := path[len("/app/"):]
	content, ok := f.files[name]
	if !ok {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	var buf bytes.Buffer
	writer := tar.NewWriter(&buf)
	assert.NoError(f.t, writer.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}))
	_, _ = writer.Write([]byte(content))
	assert.NoError(f.t, writer.Close())
	return &client.CopyFromContainerResult{Content: io.NopCloser(&buf)}, nil
}

func (f *fakeDocker) ContainerRemove(_ context.Context, _ string, options ...docker.ContainerRemoveOption) (
	*client.ContainerRemoveResult, error) {
	opts := &client.ContainerRemoveOptions{}
	for _, opt := range options {
		opt(opts)
	}
	assert.True(f.t, opts.Force)
	f.removed = true
	return &client.ContainerRemoveResult{}, nil
}

// fakeBuilder builds libraries images, or fails to.
type fakeBuilder struct {
	built []*LibrariesBuildReq
	fail  bool
	// docker gets the image once it is built.
	docker *fakeDocker
}

func (b *fakeBuilder) BuildLibraries(_ context.Context, req *LibrariesBuildReq) (string, error) {
	b.built = append(b.built, req)
	if b.fail {
		return "npm error 404 Not Found - GET https://registry.npmjs.org/nope", hperrors.Wrap(hperrors.ErrActionFailed)
	}
	b.docker.images[req.Image] = true
	return "#5 DONE 1.2s", nil
}
