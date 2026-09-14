package sysupdateserviceimpl

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	// envPGDATA is where the postgres image is told to keep the cluster.
	envPGDATA = "PGDATA"

	// dbVolumeMountTarget is where the db service mounts its volume. The image
	// keeps the cluster in a major-numbered subdirectory of it.
	dbVolumeMountTarget = "/var/lib/postgresql"

	// dbVolumeEnvFileName is where the new volume and major are recorded for the
	// next `docker stack deploy`.
	//
	// The stack file still names what it always named. Deploying it again after
	// an upgrade, without reading this, would point the service back at the old
	// volume - and at the database as it was before the upgrade. install.sh reads
	// this file; a fresh install has none.
	dbVolumeEnvFileName = "db-volume.env"

	// Readable, because install.sh sources it as the invoking user. It holds a
	// volume name and a version number, nothing secret.
	dbVolumeEnvFileMode = 0o644
)

// dbMajorUpgrade is a move of postgres from one major version to another.
//
// Postgres does not start on a data directory written by a different major, so
// this is not an image swap with a longer restart. It is: bring up an empty
// cluster of the new version on a volume of its own, load the dump taken before
// the update into it, and leave the old volume untouched as the way back.
//
// A volume of its own, and not another directory in the same one, because the
// postgres 18+ images refuse to start when they find another major's cluster in
// their volume - measured, including after renaming the directory out of the way.
// The image says so itself: "there appears to be PostgreSQL data in
// /var/lib/postgresql/18/docker ... which requires both versions" to upgrade.
type dbMajorUpgrade struct {
	CurrentImage  string
	CurrentPGDATA string
	TargetPGDATA  string
	CurrentVolume string
	TargetVolume  string
	FromMajor     int
	ToMajor       int
}

// planDbMajorUpgrade reports what a major move would take, or nil when the
// target is not one.
//
// It refuses rather than guesses. A PGDATA that does not carry the major version
// cannot be rewritten for the new one without inventing a path, and a pg_restore
// older than the server it is loading into is not something to find out halfway.
func (s *service) planDbMajorUpgrade(
	ctx context.Context,
	data *sysUpdateData,
	targetImage string,
	skipBackup bool,
) (*dbMajorUpgrade, error) {
	if targetImage == "" {
		return nil, nil
	}

	dbSvc, err := s.hpAppService.GetHpDbSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	currentImage := dbSvc.Spec.TaskTemplate.ContainerSpec.Image

	from, fromOK := imageref.MajorVersion(currentImage)
	to, toOK := imageref.MajorVersion(targetImage)
	if !fromOK || !toOK || from == to {
		return nil, nil
	}

	// The dump is not a safety net here, it is the mechanism: the new cluster
	// starts empty and the backup is the only thing that fills it. Running this
	// without one would migrate an empty database and leave an installation that
	// looks reset rather than upgraded.
	if skipBackup {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Refusing the postgres upgrade: it loads the new cluster from the backup, "+
				"and this update was asked to skip taking one",
			tasklog.TsNow))
		return nil, hperrors.Wrap(hperrors.ErrUnsupported).
			WithMsgLog("a postgres major upgrade cannot run with SkipBackup set")
	}

	currentPGDATA := pgDataEnv(dbSvc.Spec.TaskTemplate.ContainerSpec.Env)
	targetPGDATA, err := repointPGDATA(currentPGDATA, from, to)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	currentVolume, err := dbVolumeName(dbSvc.Spec.TaskTemplate.ContainerSpec.Mounts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err := s.checkRestoreClientVersion(ctx, data, to); err != nil {
		return nil, hperrors.Wrap(err)
	}

	upgrade := &dbMajorUpgrade{
		CurrentImage:  currentImage,
		CurrentPGDATA: currentPGDATA,
		TargetPGDATA:  targetPGDATA,
		CurrentVolume: currentVolume,
		TargetVolume:  repointDbVolume(currentVolume, to),
		FromMajor:     from,
		ToMajor:       to,
	}

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Upgrading postgres from major "+strconv.Itoa(from)+" to "+strconv.Itoa(to)+
			": a new cluster will be created on volume "+upgrade.TargetVolume+
			" at "+targetPGDATA+" and loaded from the backup. Volume "+currentVolume+
			" is left untouched and is what an unsuccessful upgrade returns to.",
		tasklog.TsNow))

	return upgrade, nil
}

// applyTo points the service at the new version, a new volume, and the data
// directory the new major uses. All three move together or postgres finds a
// cluster it cannot read - or refuses to start beside one.
func (u *dbMajorUpgrade) applyTo(spec *swarm.ServiceSpec) {
	spec.TaskTemplate.ContainerSpec.Env = setPGDataEnv(
		spec.TaskTemplate.ContainerSpec.Env, u.TargetPGDATA)
	setDbVolumeName(spec.TaskTemplate.ContainerSpec.Mounts, u.TargetVolume)
}

// revert puts the service back on the old version and the old cluster.
//
// This is the whole recovery path for a failed major upgrade. It works because
// nothing touched the old data directory: the new cluster was written elsewhere
// in the volume, so going back is a service update rather than a restore.
func (s *service) revertDbMajorUpgrade(
	ctx context.Context,
	data *sysUpdateData,
	upgrade *dbMajorUpgrade,
) error {
	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Putting postgres back on "+upgrade.CurrentImage+", volume "+upgrade.CurrentVolume+
			" and "+upgrade.CurrentPGDATA, tasklog.TsNow))

	dbSvc, err := s.hpAppService.GetHpDbSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, dbSvc.ID, dbSvc,
		func(_ int, svc *swarm.Service) (bool, error) {
			svc.Spec.TaskTemplate.ContainerSpec.Image = upgrade.CurrentImage
			svc.Spec.TaskTemplate.ContainerSpec.Env = setPGDataEnv(
				svc.Spec.TaskTemplate.ContainerSpec.Env, upgrade.CurrentPGDATA)
			setDbVolumeName(svc.Spec.TaskTemplate.ContainerSpec.Mounts, upgrade.CurrentVolume)
			return true, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}

	running, err := s.dockerManager.ServiceWaitUntilRunning(ctx, dbSvc.ID, true,
		dbServiceRequiredRunningDuration, serviceUpdateCheckInterval)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !running {
		return hperrors.Wrap(hperrors.ErrServiceNotRunning).WithParam("Name", "db")
	}
	return nil
}

// checkRestoreClientVersion refuses an upgrade this image cannot load.
//
// pg_restore reads forward, not backward: an 18 client cannot load a dump into a
// 19 server. The client comes from the app image, so a release that moves
// postgres has to move postgresql-client with it, and this is what says so
// before anything is changed rather than after.
func (s *service) checkRestoreClientVersion(
	ctx context.Context,
	data *sysUpdateData,
	targetMajor int,
) error {
	bin, err := exec.LookPath("pg_restore")
	if err != nil {
		return hperrors.Wrap(err)
	}

	out, err := exec.CommandContext(ctx, bin, "--version").CombinedOutput() //nolint:gosec // fixed arguments
	if err != nil {
		return hperrors.Wrap(err)
	}

	clientMajor, ok := clientMajorVersion(reflectutil.UnsafeBytesToStr(out))
	if !ok {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"Could not read a version out of `pg_restore --version`, continuing anyway", tasklog.TsNow))
		return nil
	}
	if clientMajor >= targetMajor {
		return nil
	}

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Refusing the postgres upgrade: this build ships pg_restore "+strconv.Itoa(clientMajor)+
			", which cannot load a dump into a server of major "+strconv.Itoa(targetMajor),
		tasklog.TsNow))
	return hperrors.Wrap(hperrors.ErrUnsupported).
		WithMsgLog("pg_restore is major %d, target postgres is major %d", clientMajor, targetMajor)
}

// clientMajorVersion reads `pg_restore (PostgreSQL) 18.6` as 18.
func clientMajorVersion(out string) (int, bool) {
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) == 0 {
		return 0, false
	}
	return imageref.MajorVersion("x:" + fields[len(fields)-1])
}

// repointPGDATA moves the data directory from one major's to the next one's.
//
// The major has to appear as a path segment of its own. Anything else is a path
// this cannot rewrite without making one up, and making one up here would point
// postgres at a directory nobody chose.
func repointPGDATA(current string, from, to int) (string, error) {
	if current == "" {
		return "", hperrors.Wrap(hperrors.ErrUnsupported).
			WithMsgLog("the db service sets no %s, so its data directory cannot be moved to major %d",
				envPGDATA, to)
	}

	segments := strings.Split(path.Clean(current), "/")
	replaced := false
	for i, segment := range segments {
		if segment == strconv.Itoa(from) {
			segments[i] = strconv.Itoa(to)
			replaced = true
		}
	}
	if !replaced {
		return "", hperrors.Wrap(hperrors.ErrUnsupported).
			WithMsgLog("%s is %q, which does not carry major %d as a path segment", envPGDATA, current, from)
	}
	return strings.Join(segments, "/"), nil
}

// pgDataEnv reads PGDATA out of a container spec's environment.
func pgDataEnv(env []string) string {
	for _, entry := range env {
		if value, found := strings.CutPrefix(entry, envPGDATA+"="); found {
			return value
		}
	}
	return ""
}

// setPGDataEnv replaces PGDATA, or adds it to a service that never set one.
func setPGDataEnv(env []string, value string) []string {
	prefix := envPGDATA + "="
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

// dbVolumeName reads the volume the database keeps its clusters on.
//
// A bind mount is refused rather than upgraded: an upgrade moves the database to
// a volume of its own, and there is no equivalent of that for a path an operator
// chose on the host.
func dbVolumeName(mounts []mount.Mount) (string, error) {
	for _, m := range mounts {
		if m.Target != dbVolumeMountTarget {
			continue
		}
		if m.Type != mount.TypeVolume {
			return "", hperrors.Wrap(hperrors.ErrUnsupported).
				WithMsgLog("the database is on a %s mount, which a major upgrade cannot move", m.Type)
		}
		return m.Source, nil
	}
	return "", hperrors.Wrap(hperrors.ErrUnsupported).
		WithMsgLog("the database service has no volume mounted at %s", dbVolumeMountTarget)
}

// setDbVolumeName repoints the mount, keeping its driver and labels.
func setDbVolumeName(mounts []mount.Mount, name string) {
	for i := range mounts {
		if mounts[i].Target == dbVolumeMountTarget {
			mounts[i].Source = name
		}
	}
}

// repointDbVolume names the volume for a major: hivepaas_db becomes
// hivepaas_db_19, and hivepaas_db_19 becomes hivepaas_db_20 rather than
// accumulating suffixes.
func repointDbVolume(current string, to int) string {
	if i := strings.LastIndex(current, "_"); i > 0 {
		if _, err := strconv.Atoi(current[i+1:]); err == nil {
			current = current[:i]
		}
	}
	return current + "_" + strconv.Itoa(to)
}

// recordDbVolume writes the new volume and major where install.sh will read them.
//
// Without this the next `docker stack deploy` reads the stack file, which still
// names the old volume, and silently returns the service to the database as it
// was before the upgrade.
func (s *service) recordDbVolume(ctx context.Context, data *sysUpdateData, upgrade *dbMajorUpgrade) error {
	dir := config.Current().DataPathSystemUpdate().AbsPath()
	if err := os.MkdirAll(dir, base.DirModeDefault); err != nil {
		return hperrors.Wrap(err)
	}

	path := filepath.Join(dir, dbVolumeEnvFileName)
	content := "HP_DB_VOLUME=" + upgrade.TargetVolume + "\n" +
		"HP_DB_MAJOR=" + strconv.Itoa(upgrade.ToMajor) + "\n"
	if err := os.WriteFile(path, []byte(content), dbVolumeEnvFileMode); err != nil {
		return hperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		"Recorded the new database volume in "+path+". The stack file still names the old one, "+
			"so a deploy that does not read this would put the database back as it was.",
		tasklog.TsNow))
	return nil
}
