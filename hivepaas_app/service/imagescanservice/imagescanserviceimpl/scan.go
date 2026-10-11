package imagescanserviceimpl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
)

const (
	// trivyTimeout is what one scan may take: a cold 773 MB image took 19 s at
	// 1 CPU, the Java DB a node fetches once a week 2.5 minutes.
	trivyTimeout = 10 * time.Minute
	// execTimeout is the scan's, and room for the exec around it.
	execTimeout = trivyTimeout + time.Minute

	// One scan at a time on a node: the runner's memory limit is for one, and
	// two cold scans would share it. The next waits for as long as one may take.
	nodeLockExpiry     = execTimeout + time.Minute
	nodeLockRetryDelay = 2 * time.Second
	nodeLockTries      = int(execTimeout / nodeLockRetryDelay)

	// freshScan is how long a scan stands for its image: after a deployment, or
	// the daily scan, it is not done again within it unless asked.
	freshScan = time.Hour

	// javaDBMaxAge is how old the node's Java DB may be before Trivy is let
	// renew it: it is a 935 MiB download, and Trivy's own is three days.
	javaDBMaxAgeDays = 7
)

// imageIDPattern is an image's ID as Docker gives it: what a scan's command
// line takes, and nothing else.
var imageIDPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// runningImage is an image an app's task runs, and a node it runs on.
type runningImage struct {
	ID     string
	Ref    string
	NodeID string
}

func (s *service) ScanApp(
	ctx context.Context,
	req *imagescanservice.ScanAppReq,
) (*entity.TaskImageScanOutput, error) {
	log := req.LogStore
	if log == nil {
		log = tasklog.NewNullStore()
	}
	cfg, err := s.enabled(ctx, s.db)
	if err != nil {
		return nil, err
	}
	runner, err := s.systemAppService.LoadApp(ctx, s.db, base.HivepaasTrivyRunnerKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if runner == nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
			WithExtraDetail("The runner app is not there: save the image scanning settings again.")
	}

	now := timeutil.NowUTC()
	appScan := &entity.AppImageScan{AppID: req.App.ID, ScannedAt: now, ImageIDs: []string{}}
	previous, err := s.imageScanRepo.GetAppScan(ctx, s.db, req.App.ID)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if previous != nil {
		appScan.Reported = previous.Reported
	}

	out := &entity.TaskImageScanOutput{}
	images, err := s.appImages(ctx, req.App, log)
	if err != nil {
		// What it ran last is still what there is to show, with why it was not
		// scanned now.
		if previous != nil {
			appScan.ImageIDs = previous.ImageIDs
		}
		appScan.Error = hperrors.GetErrorDetail(err, "")
		out.Error = appScan.Error
		if upsertErr := s.imageScanRepo.UpsertAppScan(ctx, s.db, appScan); upsertErr != nil {
			return out, hperrors.Wrap(upsertErr)
		}
		return out, err
	}

	ignores, err := s.ignoresOf(ctx, s.db, req.App)
	if err != nil {
		return nil, err
	}
	runners := s.runningNodes(ctx, runner)
	var failed []string
	for _, img := range images {
		result := s.scanImage(ctx, cfg, runner, runners, img, req.Force, ignores, log)
		out.Images = append(out.Images, result)
		appScan.ImageIDs = append(appScan.ImageIDs, img.ID)
		if result.Error != "" {
			failed = append(failed, fmt.Sprintf("%s: %s", shortID(img.ID), result.Error))
		}
	}
	if err = s.imageScanRepo.UpsertAppScan(ctx, s.db, appScan); err != nil {
		return out, hperrors.Wrap(err)
	}
	if len(failed) > 0 {
		return out, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%s", strings.Join(failed, "; "))
	}
	return out, nil
}

// scanImage scans one image on its node and keeps what it found, or keeps that
// it failed. An image scanned within the hour is not scanned again unless
// forced: its counts are the scan's.
func (s *service) scanImage(
	ctx context.Context,
	cfg *entity.ImageScanSettings,
	runner *entity.App,
	runners map[string]string,
	img runningImage,
	force bool,
	ignores []*imagescanservice.Ignore,
	log *tasklog.Store,
) *entity.TaskImageScanImage {
	result := &entity.TaskImageScanImage{ImageID: img.ID, ImageRef: img.Ref}
	now := timeutil.NowUTC()
	if !force {
		scans, err := s.imageScanRepo.ListScans(ctx, s.db, []string{img.ID}, true)
		if err == nil && len(scans) == 1 && scans[0].Status == base.ImageScanStatusDone &&
			now.Sub(scans[0].ScannedAt) < freshScan {
			_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("%s (%s) was scanned at %s: not again within the hour",
				img.Ref, shortID(img.ID), scans[0].ScannedAt.Format(time.RFC3339)), tasklog.TsNow))
			result.Skipped = true
			countInto(result, scans[0].Findings, ignores, now)
			return result
		}
	}

	_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Scanning %s (%s) on node %s", img.Ref, shortID(img.ID),
		img.NodeID), tasklog.TsNow))
	report, err := s.runTrivy(ctx, cfg, runner, runners[img.NodeID], img, log)
	if err != nil {
		result.Error = hperrors.GetErrorDetail(err, "")
		_ = log.Add(ctx, tasklog.NewErrFrame("Scan failed: "+result.Error, tasklog.TsNow))
		failure := &entity.ImageScan{ImageID: img.ID, ImageRef: img.Ref, Status: base.ImageScanStatusFailed,
			Error: result.Error, ScannedAt: now, UpdatedAt: now}
		if markErr := s.imageScanRepo.MarkFailed(ctx, s.db, failure); markErr != nil {
			s.logger.Errorf("image scan: failed to record the failed scan of %s: %v", img.ID, markErr)
		}
		return result
	}
	err = transaction.Execute(ctx, s.db, func(tx database.Tx) error {
		return s.imageScanRepo.ReplaceScan(ctx, tx, report.Scan, report.Findings)
	})
	if err != nil {
		result.Error = hperrors.GetErrorDetail(err, "")
		return result
	}
	countInto(result, report.Findings, ignores, now)
	_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Found %d CRITICAL and %d HIGH, %d with a fix",
		result.Critical, result.High, result.Fixable), tasklog.TsNow))
	return result
}

// countInto writes what the findings come to, after ignores, into result.
func countInto(result *entity.TaskImageScanImage, findings []*entity.ImageScanFinding,
	ignores []*imagescanservice.Ignore, at time.Time) {
	summary := imagescanservice.Summarize(findings, ignores, at)
	result.Critical, result.High = summary.Critical.Total, summary.High.Total
	result.Fixable = summary.Critical.Fixable + summary.High.Fixable + summary.Medium.Fixable +
		summary.Low.Fixable + summary.Unknown.Fixable
}

// runTrivy runs Trivy in the runner of the image's node, and reads its report.
// Its log lines go to log; its last error is the scan's.
func (s *service) runTrivy(
	ctx context.Context,
	cfg *entity.ImageScanSettings,
	runner *entity.App,
	containerID string,
	img runningImage,
	log *tasklog.Store,
) (*imagescanservice.Report, error) {
	if containerID == "" {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
			WithExtraDetail("The runner does not run on the image's node yet.")
	}
	script, err := scanScript(cfg, img.ID)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var lastErr string
	trivyLog := tasklog.NewForwardStore("", func(ctx context.Context, frames []*tasklog.LogFrame) error {
		for _, frame := range frames {
			if line := lastLine(frame.Data); line != "" && frame.Type == tasklog.LogTypeErr {
				if lastErr == "" || strings.Contains(line, "FATAL") || strings.Contains(line, "ERROR") {
					lastErr = line
				}
			}
		}
		return log.Add(ctx, frames...)
	})

	var execErr error
	lockErr := s.redisLock.Do(ctx, nodeLockKey(img.NodeID), nodeLockExpiry, nodeLockTries, nodeLockRetryDelay,
		func() {
			execCtx, cancel := context.WithTimeout(ctx, execTimeout)
			defer cancel()
			_, execErr = s.containerExecService.ContainerExec(execCtx, &containerexecservice.ContainerExecReq{
				App:         runner,
				ContainerID: containerID,
				NodeID:      img.NodeID,
				ExecOptions: func(opts *client.ExecCreateOptions) {
					opts.AttachStdout = true
					opts.AttachStderr = true
					opts.Cmd = []string{"sh", "-c", script}
					// The report is data: a TTY would not pass it through as is.
					opts.TTY = false
				},
				StdoutWriter: &stdout,
				LogStore:     trivyLog,
			})
		})
	if lockErr != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
			WithExtraDetail("Another scan of the node did not end in time: %v", lockErr)
	}
	if execErr != nil {
		if lastErr != "" {
			return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%s", trivyMessage(lastErr))
		}
		return nil, hperrors.Wrap(execErr)
	}
	report, err := imagescanservice.ParseReport(stdout.Bytes(), img.Ref, timeutil.NowUTC())
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("Trivy's report: %v", err)
	}
	return report, nil
}

// scanScript is the command a scan runs in the runner: Trivy, asking the server,
// reading the image from the node's Docker. The Java DB is renewed only once
// the node's copy is a week old.
func scanScript(cfg *entity.ImageScanSettings, imageID string) (string, error) {
	if !imageIDPattern.MatchString(imageID) {
		return "", hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%q is not an image's ID", imageID)
	}
	args := []string{
		"trivy", "image",
		"--server", serverURL(),
		"--token", fmt.Sprintf(`"$(cat %s)"`, tokenPath),
		"--cache-dir", runnerCacheDir,
		"--image-src", "docker",
		"--scanners", "vuln",
		"--list-all-pkgs=false",
		"--disable-telemetry",
		"--skip-version-check",
		"--no-progress",
		"--format", "json",
		"--timeout", trivyTimeout.String(),
		"$skip",
	}
	if cfg.JavaDBRepository != "" {
		args = append(args, "--java-db-repository", executil.ArgQuote(cfg.JavaDBRepository))
	}
	args = append(args, imageID)
	return fmt.Sprintf(`skip=""
if [ -n "$(find %s/java-db/metadata.json -mtime -%d 2>/dev/null)" ]; then skip=--skip-java-db-update; fi
exec %s`, runnerCacheDir, javaDBMaxAgeDays, strings.Join(args, " ")), nil
}

// trivyMessage is the message of a Trivy log line, without its time and level.
func trivyMessage(line string) string {
	fields := strings.SplitN(line, "\t", 3) //nolint:mnd // time, level, message
	if len(fields) == 3 {                   //nolint:mnd
		return strings.TrimSpace(fields[2])
	}
	return strings.TrimSpace(line)
}

func lastLine(data string) string {
	lines := strings.Split(strings.TrimRight(data, "\n"), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

func nodeLockKey(nodeID string) string { return "imagescan:node:lock:" + nodeID }

func shortID(imageID string) string {
	id := strings.TrimPrefix(imageID, "sha256:")
	if len(id) > 12 { //nolint:mnd // as Docker shortens one
		id = id[:12]
	}
	return id
}

// appImages are the images an app's running tasks run, one node each runs on.
// Docker is asked on each node which image a task's container was made from:
// a service carries a tag, which may since point at another image.
func (s *service) appImages(ctx context.Context, app *entity.App, log *tasklog.Store) ([]runningImage, error) {
	if app.ServiceID == "" {
		return nil, hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", app.Name).
			WithExtraDetail("The app was never deployed.")
	}
	tasks, err := s.dockerManager.ServiceTaskList(ctx, app.ServiceID, []swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	containers := map[string][]string{}
	for i := range tasks.Items {
		task := &tasks.Items[i]
		if task.Status.State != swarm.TaskStateRunning || task.Status.ContainerStatus == nil ||
			task.Status.ContainerStatus.ContainerID == "" {
			continue
		}
		containers[task.NodeID] = append(containers[task.NodeID], task.Status.ContainerStatus.ContainerID)
	}
	if len(containers) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", app.Name).
			WithExtraDetail("No task of the app is running.")
	}

	nodeIDs := make([]string, 0, len(containers))
	for nodeID := range containers {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	var images []runningImage
	seen := map[string]bool{}
	var lastErr error
	for _, nodeID := range nodeIDs {
		found, err := s.containerImages(ctx, nodeID, containers[nodeID])
		if err != nil {
			lastErr = err
			_ = log.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf("Could not read the containers of node %s: %v",
				nodeID, err), tasklog.TsNow))
			continue
		}
		for _, img := range found {
			if !seen[img.ID] {
				seen[img.ID] = true
				images = append(images, img)
			}
		}
	}
	if len(images) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return images, nil
}

// containerImages asks a node's Docker, through its agent, which images its
// containers were made from, and by what reference.
func (s *service) containerImages(ctx context.Context, nodeID string, containerIDs []string) (
	[]runningImage, error) {
	var stdout, stderr bytes.Buffer
	command := append([]string{"docker", "inspect", "--format", "{{.Image}} {{.Config.Image}}"}, containerIDs...)
	resp, err := s.nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
		NodeID:          nodeID,
		CommandExecOpts: &nodeexecservice.CommandExecOpts{Command: command, Stdout: &stdout, Stderr: &stderr},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	images := parseContainerImages(stdout.String(), nodeID)
	if resp.ExitCode != 0 && len(images) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).WithExtraDetail("docker inspect: %s",
			strings.TrimSpace(stderr.String()))
	}
	return images, nil
}

// parseContainerImages reads `docker inspect` lines of an image's ID and the
// reference the container was made with.
func parseContainerImages(out, nodeID string) []runningImage {
	var images []runningImage
	for _, line := range strings.Split(out, "\n") {
		id, ref, _ := strings.Cut(strings.TrimSpace(line), " ")
		if !imageIDPattern.MatchString(id) {
			continue
		}
		images = append(images, runningImage{ID: id, Ref: imageWithoutDigest(ref), NodeID: nodeID})
	}
	return images
}

// ignoresOf are the ignores that apply to an app.
func (s *service) ignoresOf(ctx context.Context, db database.IDB, app *entity.App) (
	[]*imagescanservice.Ignore, error) {
	settings, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeVulnerabilityIgnore),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return imagescanservice.IgnoresOf(app, settings), nil
}
