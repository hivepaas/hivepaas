package imagescanserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/distribution/reference"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
)

const (
	// trivyTimeout is what one scan may take: a cold 773 MB image took 19 s at
	// 1 CPU, the Java DB a volume gets once a week 2.5 minutes.
	trivyTimeout = 10 * time.Minute
	// callTimeout is a scan's on its node: Trivy's, its DBs' version, the
	// container's start and stop, and Trivy's image pulled when the node lacks
	// it.
	callTimeout = trivyTimeout + 3*time.Minute

	// One scan at a time per copy of the volume: Trivy's cache is a file one
	// process opens at a time. The next waits for as long as one may take.
	lockExpiry     = callTimeout + time.Minute
	lockRetryDelay = 2 * time.Second
	lockTries      = int(callTimeout / lockRetryDelay)

	// freshScan is how long a scan stands for its image: after a deployment, or
	// the daily scan, it is not done again within it unless asked.
	freshScan = time.Hour
)

// imageIDPattern is an image's ID as Docker gives it.
var imageIDPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// runningImage is an image an app's task runs, and the node it runs on.
type runningImage struct {
	ID     string
	Ref    string
	NodeID string
	// Platform is the node's: os/arch, what is read of an image of several
	// platforms. Docker 29's image inspect names no variant when there is none,
	// and its template fails on one it does not name: arm64's is v8, the one
	// Trivy reads for linux/arm64.
	Platform string
	// RemoteRef is the image's repository and digest, where a node that does not
	// have it reads it; empty when it is in no registry.
	RemoteRef string
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
	volume, err := s.loadVolume(ctx, s.db, cfg)
	if err != nil {
		return nil, err
	}
	images, imagesErr := s.appImages(ctx, req.App, log)
	ignores, err := s.ignoresOf(ctx, s.db, req.App)
	if err != nil {
		return nil, err
	}
	auth, err := s.pullAuth(ctx, req.App)
	if err != nil {
		_ = log.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf("The app's registry credentials could not be read, "+
			"an image read from its registry is read without: %s", hperrors.GetErrorDetail(err, "")), tasklog.TsNow))
	}
	return s.scanSubject(ctx, &subject{id: req.App.ID, images: images, imagesErr: imagesErr, ignores: ignores,
		auth: auth}, &imageScanReq{cfg: cfg, volume: volume, force: req.Force, log: log})
}

// subject is what a scan is of: an app, or a part of HivePaaS.
type subject struct {
	// id is the app's ID, or the component's.
	id string
	// images are those it runs; imagesErr why they could not be read.
	images    []runningImage
	imagesErr error
	ignores   []*imagescanservice.Ignore
	// auth is the credential its images are pulled with; nil for none.
	auth *registry.AuthConfig
}

// scanSubject scans the images a subject runs, with the settings, the volume,
// the force and the log of with, and keeps what it ran: one whose images could
// not be read keeps what it ran last, with why.
func (s *service) scanSubject(ctx context.Context, sub *subject, with *imageScanReq) (
	*entity.TaskImageScanOutput, error) {
	appScan := &entity.AppImageScan{AppID: sub.id, ScannedAt: timeutil.NowUTC(), ImageIDs: []string{}}
	previous, err := s.imageScanRepo.GetAppScan(ctx, s.db, sub.id)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if previous != nil {
		appScan.Reported = previous.Reported
	}

	out := &entity.TaskImageScanOutput{}
	if sub.imagesErr != nil {
		if previous != nil {
			appScan.ImageIDs = previous.ImageIDs
		}
		appScan.Error = hperrors.GetErrorDetail(sub.imagesErr, "")
		out.Error = appScan.Error
		if upsertErr := s.imageScanRepo.UpsertAppScan(ctx, s.db, appScan); upsertErr != nil {
			return out, hperrors.Wrap(upsertErr)
		}
		return out, sub.imagesErr
	}

	var failed []string
	for _, img := range sub.images {
		req := *with
		req.img, req.auth, req.ignores = img, sub.auth, sub.ignores
		result := s.scanImage(ctx, &req)
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

// imageScanReq is one image's scan, and what it is scanned by.
type imageScanReq struct {
	cfg     *entity.ImageScanSettings
	volume  *scanVolume
	auth    *registry.AuthConfig
	img     runningImage
	force   bool
	ignores []*imagescanservice.Ignore
	log     *tasklog.Store
}

// scanImage scans one image and keeps what it found, or keeps that it failed.
// An image scanned within the hour is not scanned again unless forced: its
// counts are the scan's.
func (s *service) scanImage(ctx context.Context, req *imageScanReq) *entity.TaskImageScanImage {
	img, log := req.img, req.log
	result := &entity.TaskImageScanImage{ImageID: img.ID, ImageRef: img.Ref}
	now := timeutil.NowUTC()
	if !req.force {
		scans, err := s.imageScanRepo.ListScans(ctx, s.db, []string{img.ID}, true)
		if err == nil && len(scans) == 1 && scans[0].Status == base.ImageScanStatusDone &&
			now.Sub(scans[0].ScannedAt) < freshScan {
			_ = log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("%s (%s) was scanned at %s: not again within the hour",
				img.Ref, shortID(img.ID), scans[0].ScannedAt.Format(time.RFC3339)), tasklog.TsNow))
			result.Skipped = true
			countInto(result, scans[0].Findings, req.ignores, now)
			return result
		}
	}

	report, err := s.runScan(ctx, req)
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
	countInto(result, report.Findings, req.ignores, now)
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

// runScan runs the image's scan on the node the volume decides, one at a time
// per copy of the volume, and reads what it found. Trivy's log goes to the
// task's.
func (s *service) runScan(ctx context.Context, req *imageScanReq) (*imagescanservice.Report, error) {
	img, cfg := req.img, req.cfg
	scanNode := req.volume.scanNode(img.NodeID)
	auth, err := dockerConfig(req.auth, img.RemoteRef)
	if err != nil {
		return nil, err
	}
	runReq := &imagescanrun.RunReq{
		TrivyImage:       trivyImage(),
		ImageID:          img.ID,
		ImageNode:        img.NodeID,
		RemoteRef:        img.RemoteRef,
		Platform:         img.Platform,
		DockerConfig:     auth,
		Volume:           req.volume.Mount,
		CacheDir:         cacheDir(),
		DBRepository:     cfg.DBRepository,
		JavaDBRepository: cfg.JavaDBRepository,
		NanoCPUs:         int64(cfg.ScanCPULimit() * 1e9), //nolint:mnd // CPUs in billionths
		MemoryBytes:      cfg.ScanMemoryLimit().Bytes(),
		Timeout:          trivyTimeout,
	}
	_ = req.log.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Scanning %s (%s), run on node %s, on node %s",
		img.Ref, shortID(img.ID), img.NodeID, scanNode), tasklog.TsNow))

	resp, err := s.runLocked(ctx, req.volume, scanNode, runReq)
	if err != nil {
		return nil, err
	}
	if resp.Log != "" {
		_ = req.log.Add(ctx, tasklog.NewOutFrame(strings.TrimRight(resp.Log, "\n"), tasklog.TsNow))
	}
	if resp.Source == base.ImageScanSourceRegistry {
		_ = req.log.Add(ctx, tasklog.NewOutFrame("Read from its registry: "+img.RemoteRef, tasklog.TsNow))
	}
	return readScan(resp, img, cfg, timeutil.NowUTC())
}

// runLocked runs a scan on its node, once the volume's copy there is free.
func (s *service) runLocked(ctx context.Context, volume *scanVolume, scanNode string,
	runReq *imagescanrun.RunReq) (*imagescanrun.RunResp, error) {
	var resp *imagescanrun.RunResp
	var runErr error
	lockErr := s.redisLock.Do(ctx, volume.lockKey(scanNode), lockExpiry, lockTries, lockRetryDelay, func() {
		callCtx, cancel := context.WithTimeout(ctx, callTimeout)
		defer cancel()
		resp, runErr = s.scanOn(callCtx, scanNode, runReq)
	})
	if lockErr != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
			WithExtraDetail("Another scan on the volume did not end in time: %v", lockErr)
	}
	return resp, runErr
}

// scanOn runs a scan on a node: in this process on its own, through the node's
// agent on another.
func (s *service) scanOn(ctx context.Context, nodeID string, req *imagescanrun.RunReq) (
	*imagescanrun.RunResp, error) {
	current, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	remote := nodeID != current
	if cfg := config.Current(); cfg != nil && cfg.DevMode.Enabled && cfg.DevMode.ForceAgentLocal {
		remote = true
	}
	if !remote {
		resp, err := s.localScanner.Run(ctx, req)
		return resp, hperrors.Wrap(err)
	}
	addr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanNotReady).
			WithExtraDetail("The agent of node %s cannot be reached: %v", nodeID, err)
	}
	agent, err := s.newAgentClient(addr)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = agent.Close() }()
	resp, err := agent.Run(ctx, req)
	return resp, hperrors.Wrap(err)
}

// readScan reads what a scan's container printed into the image's scan: by
// the image's ID on the node that runs it, where it was read, and the DB's
// date. A container that printed no report failed with Trivy's last error.
func readScan(resp *imagescanrun.RunResp, img runningImage, cfg *entity.ImageScanSettings, at time.Time) (
	*imagescanservice.Report, error) {
	if resp.OOMKilled {
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail(
			"The scan ran out of its memory limit, %v: raise it in System > Image Scanning.",
			cfg.ScanMemoryLimit())
	}
	out, err := imagescanservice.ParseScanOutput(resp.Output)
	if err != nil || !out.HasReport() {
		if line := lastErrorLine(resp.Log); line != "" {
			return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("%s", trivyMessage(line))
		}
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).
			WithExtraDetail("Trivy stopped with exit code %d, and said nothing more.", resp.ExitCode)
	}
	report, err := imagescanservice.ParseReport(out.Report, img.ID, img.Ref, at)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrImageScanFailed).WithExtraDetail("Trivy's report: %v", err)
	}
	report.Scan.Source = resp.Source
	if updatedAt := out.DBUpdatedAt(); !updatedAt.IsZero() {
		report.Scan.DBUpdatedAt = updatedAt
	}
	return report, nil
}

// lastErrorLine is the last FATAL or ERROR line of Trivy's log, or its last
// line when none is.
func lastErrorLine(log string) string {
	lines := strings.Split(strings.TrimRight(log, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "\tFATAL\t") || strings.Contains(lines[i], "\tERROR\t") {
			// A fatal error's message goes on over the lines after it.
			return strings.Join(lines[i:], "\n")
		}
	}
	return strings.TrimSpace(lines[len(lines)-1])
}

// trivyMessage is the message of a Trivy log line, without its time and level.
func trivyMessage(line string) string {
	fields := strings.SplitN(line, "\t", 3) //nolint:mnd // time, level, message
	if len(fields) == 3 {                   //nolint:mnd
		return strings.TrimSpace(fields[2])
	}
	return strings.TrimSpace(line)
}

func shortID(imageID string) string {
	id := strings.TrimPrefix(imageID, "sha256:")
	if len(id) > 12 { //nolint:mnd // as Docker shortens one
		id = id[:12]
	}
	return id
}

// appImages are the images an app's running tasks run, one node each runs on.
func (s *service) appImages(ctx context.Context, app *entity.App, log *tasklog.Store) ([]runningImage, error) {
	if app.ServiceID == "" {
		return nil, hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", app.Name).
			WithExtraDetail("The app was never deployed.")
	}
	return s.serviceImages(ctx, app.ServiceID, app.Name, log)
}

// serviceImages are the images a service's running tasks run, one node each
// runs on. Docker is asked on each node which image a task's container was
// made from: a service carries a tag, which may since point at another image.
func (s *service) serviceImages(ctx context.Context, serviceID, name string, log *tasklog.Store) (
	[]runningImage, error) {
	tasks, err := s.dockerManager.ServiceTaskList(ctx, serviceID, []swarm.TaskState{swarm.TaskStateRunning})
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
		return nil, hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", name).
			WithExtraDetail("No task of %s is running.", name)
	}
	return s.imagesOnNodes(ctx, containers, log)
}

// imagesOnNodes are the images the containers run, by node, one node each
// runs on: the first by ID.
func (s *service) imagesOnNodes(ctx context.Context, containers map[string][]string, log *tasklog.Store) (
	[]runningImage, error) {
	nodeIDs := make([]string, 0, len(containers))
	for nodeID := range containers {
		nodeIDs = append(nodeIDs, nodeID)
	}
	slices.Sort(nodeIDs)
	var images []runningImage
	seen := map[string]bool{}
	var lastErr error
	for _, nodeID := range nodeIDs {
		found, err := s.containerImages(ctx, nodeID, containers[nodeID], log)
		if err != nil {
			lastErr = err
			_ = log.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf("Could not read the containers of node %s: %s",
				nodeID, hperrors.GetErrorDetail(err, "")), tasklog.TsNow))
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
// containers were made from and by what reference, and of each image the
// node's platform and the digests it was pulled or pushed by.
func (s *service) containerImages(ctx context.Context, nodeID string, containerIDs []string,
	log *tasklog.Store) ([]runningImage, error) {
	stdout, err := s.dockerOnNode(ctx, nodeID,
		append([]string{"docker", "inspect", "--format", "{{.Image}} {{.Config.Image}}"}, containerIDs...))
	if err != nil {
		return nil, err
	}
	images := parseContainerImages(stdout, nodeID)
	if len(images) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(images))
	for _, img := range images {
		ids = append(ids, img.ID)
	}
	stdout, err = s.dockerOnNode(ctx, nodeID, append([]string{"docker", "image", "inspect", "--format",
		"{{.Id}} {{.Os}}/{{.Architecture}} {{json .RepoDigests}}"}, ids...))
	if err != nil {
		// Read on its node, an image needs neither.
		_ = log.Add(ctx, tasklog.NewWarnFrame(fmt.Sprintf("Could not read the images of node %s: %s",
			nodeID, hperrors.GetErrorDetail(err, "")), tasklog.TsNow))
		return images, nil
	}
	details := parseImageDetails(stdout)
	for i := range images {
		if detail, ok := details[images[i].ID]; ok {
			images[i].Platform = detail.platform
			images[i].RemoteRef = repoDigestFor(images[i].Ref, detail.repoDigests)
		}
	}
	return images, nil
}

// dockerOnNode runs a docker command on a node, through its agent, and answers
// what it printed. It fails only when the command printed nothing.
func (s *service) dockerOnNode(ctx context.Context, nodeID string, command []string) (string, error) {
	var stdout, stderr bytes.Buffer
	resp, err := s.nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
		NodeID:          nodeID,
		CommandExecOpts: &nodeexecservice.CommandExecOpts{Command: command, Stdout: &stdout, Stderr: &stderr},
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if resp.ExitCode != 0 && strings.TrimSpace(stdout.String()) == "" {
		return "", hperrors.Wrap(hperrors.ErrImageScanNotReady).WithExtraDetail("%s: %s",
			strings.Join(command[:2], " "), strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
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

// imageDetail is what a node says of an image it has.
type imageDetail struct {
	platform    string
	repoDigests []string
}

// parseImageDetails reads `docker image inspect` lines of an image's ID, its
// platform and its repo digests as JSON.
func parseImageDetails(out string) map[string]imageDetail {
	details := map[string]imageDetail{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), " ", 3) //nolint:mnd // id, platform, digests
		if len(fields) < 2 || !imageIDPattern.MatchString(fields[0]) {
			continue
		}
		detail := imageDetail{platform: fields[1]}
		if len(fields) == 3 { //nolint:mnd
			_ = json.Unmarshal([]byte(fields[2]), &detail.repoDigests)
		}
		details[fields[0]] = detail
	}
	return details
}

// repoDigestFor is the digest of ref's repository among an image's repo
// digests, or the first of them when none is that repository's: an image is
// the same content whatever name it is read by. Empty when it has none.
func repoDigestFor(ref string, repoDigests []string) string {
	want := ""
	if named, err := reference.ParseNormalizedNamed(ref); err == nil {
		want = named.Name()
	}
	first := ""
	for _, digest := range repoDigests {
		named, err := reference.ParseNormalizedNamed(digest)
		if err != nil {
			continue
		}
		if _, canonical := named.(reference.Canonical); !canonical {
			continue
		}
		if named.Name() == want {
			return digest
		}
		if first == "" {
			first = digest
		}
	}
	return first
}

// imageWithoutDigest is a reference without its digest: the name and tag it
// is shown by.
func imageWithoutDigest(ref string) string {
	if at := strings.LastIndex(ref, "@"); at > 0 {
		return ref[:at]
	}
	return ref
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
