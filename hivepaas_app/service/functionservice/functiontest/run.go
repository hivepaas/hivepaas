package functiontest

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/services/docker"
)

// removeTimeout bounds the removal of a run's container, which happens however
// the run ends, a canceled one included.
const removeTimeout = 30 * time.Second

// Runner runs test runs on this node.
type Runner struct {
	Docker  docker.Manager
	Builder LibrariesBuilder
	// TempDir is where a run writes the function's files for its libraries'
	// build; the run removes what it wrote.
	TempDir string
	// KillMargin replaces how long a container may outlive the call's timeout;
	// zero keeps the runtime's.
	KillMargin time.Duration
}

// Run calls the function once: its libraries' image built or found, a
// container made from it with the code and the request copied in, invoke run
// in it, what it wrote read, and the container removed.
func (r *Runner) Run(ctx context.Context, req *RunReq) (*RunResp, error) {
	dir, err := fileutil.CreateTempDir(r.TempDir, "function-test-*", 0o700) //nolint:mnd
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer os.RemoveAll(dir)
	contextDir := filepath.Join(dir, "code")
	if err = os.Mkdir(contextDir, 0o700); err != nil { //nolint:mnd
		return nil, hperrors.Wrap(err)
	}
	if err = functionbuild.WriteInlineCode(contextDir, &entity.FunctionInlineCode{Files: req.Files}); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &RunResp{}
	image, err := r.librariesImage(ctx, req, contextDir, dir, resp)
	if err != nil || resp.Outcome != "" {
		return resp, err
	}

	created, err := r.Docker.ContainerCreate(ctx, func(opts *client.ContainerCreateOptions) {
		*opts = containerOptions(req, image)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() {
		removeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), removeTimeout)
		defer cancel()
		_, _ = r.Docker.ContainerRemove(removeCtx, created.ID, func(opts *client.ContainerRemoveOptions) {
			opts.Force = true
		})
	}()

	if err = r.copyIn(ctx, created.ID, req); err != nil {
		return nil, err
	}
	if _, err = r.Docker.ContainerStart(ctx, created.ID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	exitCode, killed, err := r.wait(ctx, created.ID, req)
	if err != nil {
		return nil, err
	}
	if err = r.readOutput(ctx, created.ID, resp); err != nil {
		return nil, err
	}
	resp.ExitCode = exitCode
	switch {
	case killed:
		resp.Outcome = OutcomeKilled
	case exitCode == 2: //nolint:mnd // the contract's: the request could not be read
		resp.Outcome = OutcomeBadRequest
	case exitCode == 3: //nolint:mnd // the contract's: the handler could not be loaded or built
		resp.Outcome = OutcomeNotLoaded
	case resp.Outcome == "":
		resp.Outcome = OutcomeNoResult
	}
	if resp.LockFiles, err = r.lockFiles(ctx, created.ID, req); err != nil {
		return nil, err
	}
	return resp, nil
}

// librariesImage is the image the run starts from: the libraries' image, built
// when this node does not have it, or the runtime's, pulled when it does not.
// A build that fails ends the run, its outcome in resp.
func (r *Runner) librariesImage(
	ctx context.Context,
	req *RunReq,
	contextDir, tempDir string,
	resp *RunResp,
) (string, error) {
	libs, err := functionbuild.Libraries(&functionbuild.DockerfileReq{
		Source:       req.Source,
		Images:       req.Images,
		SourceDir:    contextDir,
		BuildArgs:    gofn.MapKeys(req.Inputs.EnvVars),
		BuildSecrets: gofn.MapKeys(req.Inputs.SecretEnvVars),
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	found, err := r.hasImage(ctx, libs.Image)
	if err != nil || found {
		return libs.Image, err
	}

	if libs.Dockerfile == "" {
		pull, err := r.Docker.ImagePull(ctx, libs.Image)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		defer pull.Close()
		return libs.Image, hperrors.Wrap(pull.Wait(ctx))
	}

	log, err := r.Builder.BuildLibraries(ctx, &LibrariesBuildReq{
		Image:         libs.Image,
		Dockerfile:    libs.Dockerfile,
		ContextDir:    contextDir,
		TempDir:       tempDir,
		Inputs:        req.Inputs,
		BuildSettings: req.BuildSettings,
	})
	resp.LibrariesBuilt, resp.LibrariesLog = true, log
	if err != nil {
		if ctx.Err() != nil {
			return "", hperrors.Wrap(ctx.Err())
		}
		resp.Outcome = OutcomeLibrariesFailed
		resp.Error = err.Error()
	}
	return libs.Image, nil
}

func (r *Runner) hasImage(ctx context.Context, image string) (bool, error) {
	_, err := r.Docker.ImageInspect(ctx, image)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, hperrors.ErrInfraNotFound):
		return false, nil
	}
	return false, hperrors.Wrap(err)
}

// containerOptions is the run's container: the function's environment with its
// limits, its network and its resources, and no published port. Its name and
// labels make it a temporary one, which the cluster cleanup removes should this
// run not.
func containerOptions(req *RunReq, image string) client.ContainerCreateOptions {
	return client.ContainerCreateOptions{
		Name: docker.TempContainerPrefix + "fn-" + gofn.RandString(8), //nolint:mnd
		Config: &container.Config{
			Image: image,
			Cmd:   []string{"sh", "-c", "exec hivepaas-runtime invoke < " + requestPath},
			Env:   containerEnv(req),
			Labels: map[string]string{
				docker.LabelTempResource:  docker.LabelTempResourceVal,
				docker.LabelTempCreatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
		HostConfig: &container.HostConfig{
			NetworkMode: container.NetworkMode(req.Network),
			Resources: container.Resources{
				NanoCPUs: req.NanoCPUs,
				Memory:   req.MemoryBytes,
			},
		},
	}
}

// containerEnv is the function's environment, then its entrypoint and limits:
// whatever its environment says of those, its settings decide.
func containerEnv(req *RunReq) []string {
	src := req.Source
	env := slices.DeleteFunc(slices.Clone(req.Env), func(v string) bool {
		return strings.HasPrefix(v, "HP_FN_")
	})
	return append(env,
		"HP_FN_ENTRYPOINT="+src.Entrypoint.File,
		"HP_FN_HANDLER="+src.Entrypoint.Handler,
		fmt.Sprintf("HP_FN_TIMEOUT_MS=%d", time.Duration(src.Timeout).Milliseconds()),
		fmt.Sprintf("HP_FN_MAX_CONCURRENCY=%d", src.MaxConcurrency),
		fmt.Sprintf("HP_FN_MAX_BODY_SIZE=%d", src.MaxBodySize.Bytes()),
	)
}

// copyIn copies the code into /app and the request where the command reads it,
// both owned by the runtime's user.
func (r *Runner) copyIn(ctx context.Context, containerID string, req *RunReq) error {
	code, err := tarOf(req.Files)
	if err != nil {
		return err
	}
	_, err = r.Docker.ContainerCopyTo(ctx, containerID, "/app", code, docker.ContainerCopyToWithCopyUIDGID(true))
	if err != nil {
		return hperrors.Wrap(err)
	}

	request, err := json.Marshal(req.Request)
	if err != nil {
		return hperrors.Wrap(err)
	}
	content, err := tarOf([]*entity.FunctionFile{{Path: filepath.Base(requestPath), Content: string(request)}})
	if err != nil {
		return err
	}
	_, err = r.Docker.ContainerCopyTo(ctx, containerID, filepath.Dir(requestPath), content,
		docker.ContainerCopyToWithCopyUIDGID(true))
	return hperrors.Wrap(err)
}

// tarOf is files as a tar, owned by the runtime's user.
func tarOf(files []*entity.FunctionFile) (io.Reader, error) {
	var buf bytes.Buffer
	writer := tar.NewWriter(&buf)
	dirs := map[string]bool{}
	for _, file := range files {
		for dir := filepath.Dir(file.Path); dir != "." && !dirs[dir]; dir = filepath.Dir(dir) {
			dirs[dir] = true
		}
	}
	for _, dir := range slices.Sorted(maps.Keys(dirs)) {
		err := writer.WriteHeader(&tar.Header{Typeflag: tar.TypeDir, Name: dir + "/", Mode: 0o755, //nolint:mnd
			Uid: RuntimeUID, Gid: RuntimeUID})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	for _, file := range files {
		err := writer.WriteHeader(&tar.Header{Typeflag: tar.TypeReg, Name: file.Path, Mode: 0o644, //nolint:mnd
			Size: int64(len(file.Content)), Uid: RuntimeUID, Gid: RuntimeUID})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if _, err = writer.Write([]byte(file.Content)); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &buf, hperrors.Wrap(writer.Close())
}

// wait waits for the container to exit, and kills it when it outlives the
// call's timeout and its margin.
func (r *Runner) wait(ctx context.Context, containerID string, req *RunReq) (exitCode int64, killed bool, _ error) {
	margin := killMargin
	if req.Source.Runtime.Compiled() {
		margin = compiledKillMargin
	}
	if r.KillMargin > 0 {
		margin = r.KillMargin
	}
	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Source.Timeout)+margin)
	defer cancel()

	waited := r.Docker.ContainerWait(waitCtx, containerID, func(opts *client.ContainerWaitOptions) {
		opts.Condition = container.WaitConditionNotRunning
	})
	select {
	case result := <-waited.Result:
		return result.StatusCode, false, nil
	case err := <-waited.Error:
		if waitCtx.Err() == nil {
			return 0, false, hperrors.Wrap(err)
		}
	case <-waitCtx.Done():
	}
	if ctx.Err() != nil {
		return 0, false, hperrors.Wrap(ctx.Err())
	}
	if _, err := r.Docker.ContainerKill(ctx, containerID, "KILL"); err != nil {
		return 0, false, hperrors.Wrap(err)
	}
	return 0, true, nil
}

// readOutput reads what the container wrote: the result, the call's log and
// its standard error.
func (r *Runner) readOutput(ctx context.Context, containerID string, resp *RunResp) error {
	logs, err := r.Docker.ContainerLogs(ctx, containerID, func(opts *client.ContainerLogsOptions) {
		opts.ShowStdout, opts.ShowStderr = true, true
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer logs.Close()
	var stdout, stderr bytes.Buffer
	if _, err = stdcopy.StdCopy(&stdout, &stderr, logs); err != nil {
		return hperrors.Wrap(err)
	}
	resp.Error = stderr.String()

	out := stdout.String()
	if i := strings.LastIndex(out, ResultMarker); i >= 0 && (i == 0 || out[i-1] == '\n') {
		if err = readResult(strings.TrimSpace(out[i+len(ResultMarker):]), resp); err != nil {
			return err
		}
		out = out[:i]
	}
	if len(out) > int(LogsMax) {
		out, resp.LogsTruncated = out[len(out)-int(LogsMax):], true
	}
	resp.Logs = out
	return nil
}

func readResult(line string, resp *RunResp) error {
	var result struct {
		Status     int                 `json:"status"`
		Headers    map[string][]string `json:"headers"`
		Body       []byte              `json:"body"`
		RequestID  string              `json:"requestId"`
		DurationMs float64             `json:"durationMs"`
		Outcome    Outcome             `json:"outcome"`
	}
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return hperrors.Wrap(err)
	}
	resp.Outcome, resp.Status, resp.Headers = result.Outcome, result.Status, result.Headers
	resp.RequestID, resp.DurationMs = result.RequestID, result.DurationMs
	resp.Body = result.Body
	if len(resp.Body) > int(BodyMax) {
		resp.Body, resp.BodyTruncated = resp.Body[:BodyMax], true
	}
	return nil
}

// lockFiles are the runtime's lock files the run made or changed: the ones the
// code sent did not have, or had otherwise.
func (r *Runner) lockFiles(ctx context.Context, containerID string, req *RunReq) ([]*entity.FunctionFile, error) {
	var out []*entity.FunctionFile
	for _, name := range functionbuild.LockFiles(req.Source.Runtime) {
		content, found, err := r.readFile(ctx, containerID, "/app/"+name)
		if err != nil {
			return nil, err
		}
		sent, _ := gofn.Find(req.Files, func(f *entity.FunctionFile) bool { return f.Path == name })
		if found && (sent == nil || sent.Content != content) {
			out = append(out, &entity.FunctionFile{Path: name, Content: content})
		}
	}
	return out, nil
}

func (r *Runner) readFile(ctx context.Context, containerID, path string) (string, bool, error) {
	copied, err := r.Docker.ContainerCopyFrom(ctx, containerID, path)
	if errors.Is(err, hperrors.ErrInfraNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, hperrors.Wrap(err)
	}
	defer copied.Content.Close()
	reader := tar.NewReader(copied.Content)
	if _, err = reader.Next(); err != nil {
		return "", false, hperrors.Wrap(err)
	}
	// A lock file is code the editor adds: it has the inline code's limit.
	content, err := io.ReadAll(io.LimitReader(reader, base.FunctionInlineCodeMaxSize.Bytes()))
	if err != nil {
		return "", false, hperrors.Wrap(err)
	}
	return string(content), true, nil
}
