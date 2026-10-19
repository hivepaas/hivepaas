package postgres

import (
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// What a restored app is restored with, and what walg-entrypoint reads it from,
// as engine API 1's contract (design §4, "The steps"):
//
//   - The source's archive is told as the app's own is, each variable's name
//     prefixed with restoreEnvPrefix, and its keys are files of
//     restoreSecretsDir, with their usual names. walg-fetch reads the source
//     with those in place of the app's own; nothing writes to it.
//   - envRestoreBackup is the base backup to fetch: a name, or LATEST. Without
//     it, envRestoreBefore is a time, and the backup fetched is the latest that
//     starts before it.
//   - envRestoreRecovery is the recovery settings to write, space-separated
//     name=value pairs; recovery_target_action=promote is always added.
//
// Every variable of the restore starts with restoreEnvPrefix, and every secret
// with restoreSecretPrefix: that is how FinishRestore finds them.
const (
	restoreEnvPrefix     = "HIVEPAAS_RESTORE_"
	envRestoreSecretsDir = restoreEnvPrefix + "SECRETS_DIR"
	envRestoreBackup     = restoreEnvPrefix + "TARGET_BACKUP"
	envRestoreBefore     = restoreEnvPrefix + "TARGET_BEFORE"
	envRestoreRecovery   = restoreEnvPrefix + "RECOVERY"
	restoreSecretPrefix  = "RESTORE_"
	restoreSecretsDir    = walgkit.SecretsDir + "/restore" //nolint:gosec // G101: a directory
	// restoreSourceDir is where a volume archive the app is restored from is
	// mounted, read-only.
	restoreSourceDir = "/var/lib/hivepaas-walg-source"
	// recoveryLatest is backup-fetch's name for the newest base backup.
	recoveryLatest = "LATEST"
	// recoveryTimeLayout is a time recovery_target_time takes: ISO 8601 with a
	// numeric offset. Postgres refuses the RFC 3339 Z, as it does any zone name.
	recoveryTimeLayout = "2006-01-02T15:04:05-07:00"
)

var (
	// lsnPattern and backupPattern are what a target may say, before it reaches
	// the server's configuration or WAL-G's arguments.
	lsnPattern    = regexp.MustCompile(`^[0-9A-Fa-f]{1,8}/[0-9A-Fa-f]{1,8}$`)
	backupPattern = regexp.MustCompile(`^base_[0-9A-Fa-f]{24}(_D_[0-9A-Fa-f]{24})?$`)
)

// prepareRestore adds to a restored app's document the source's archive and its
// keys, and the target, and returns the variables to add.
func (e *Engine) prepareRestore(doc *specmodel.AppDoc, source *managedmodel.RestoreSource) ([][2]string, error) {
	target, err := targetEnv(source)
	if err != nil {
		return nil, err
	}
	env := [][2]string{{envRestoreSecretsDir, restoreSecretsDir}}
	for _, pair := range walgkit.ArchiveEnv(source.Archive, restoreSourceDir) {
		env = append(env, [2]string{restoreEnvPrefix + pair[0], pair[1]})
	}
	if source.Archive.Volume != nil {
		walgkit.MountArchiveVolume(doc, restoreSourceDir, source.Archive.Volume, true)
	}
	secrets := e.ArchiveSecrets(source.Archive)
	secrets[walgkit.SecretSodiumKey] = source.EncryptionKey
	walgkit.AddSecretFiles(doc, restoreSecretsDir, restoreSecretPrefix, postgresUID, secrets)
	return append(env, target...), nil
}

// targetEnv is a target as what to fetch and how to recover (design §4, "What
// to restore to").
func targetEnv(source *managedmodel.RestoreSource) ([][2]string, error) {
	target := source.Target
	switch target.Kind {
	case entity.ManagedRestoreToLatest:
		return [][2]string{{envRestoreBackup, recoveryLatest}}, nil
	case entity.ManagedRestoreToTime:
		if target.Time.IsZero() {
			return nil, targetInvalid("a time is required")
		}
		at := target.Time.UTC()
		return [][2]string{{envRestoreBefore, at.Format(time.RFC3339)},
			{envRestoreRecovery, "recovery_target_time=" + at.Format(recoveryTimeLayout)}}, nil
	case entity.ManagedRestoreToPoint:
		if !lsnPattern.MatchString(target.Position) || source.PointAt.IsZero() {
			return nil, targetInvalid("the point has no LSN to recover to")
		}
		return [][2]string{
			{envRestoreBefore, source.PointAt.UTC().Format(time.RFC3339)},
			{envRestoreRecovery, "recovery_target_lsn=" + target.Position + " recovery_target_inclusive=on"},
		}, nil
	case entity.ManagedRestoreToBackup:
		if !backupPattern.MatchString(target.Backup) {
			return nil, targetInvalid("not a base backup's name")
		}
		return [][2]string{{envRestoreBackup, target.Backup}, {envRestoreRecovery, "recovery_target=immediate"}}, nil
	}
	return nil, targetInvalid("the target is latest, time, point or backup")
}

func targetInvalid(detail string) error {
	return hperrors.Wrap(hperrors.ErrAppManagedRestoreTargetInvalid).WithExtraDetail("%s", detail)
}

func (e *Engine) RecoveryCheck() managedmodel.RecoveryCheck {
	return managedmodel.RecoveryCheck{Command: walgkit.ScriptDir + "/walg-entrypoint status"}
}

func (e *Engine) FirstBackupJob(managed *entity.AppManagedSettings) string {
	if managed == nil || managed.Database == nil {
		return ""
	}
	return managed.Database.Jobs.Nightly.ID
}

func (e *Engine) FinishRestore(req *managedmodel.FinishRestoreReq) (*managedmodel.FinishRestoreResp, error) {
	pg := req.Managed.Database
	if pg == nil || pg.Restore == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("postgres: not a restored database")
	}
	resp := &managedmodel.FinishRestoreResp{}
	deleted := map[string]bool{}
	for _, setting := range req.Settings {
		switch setting.Type { //nolint:exhaustive // the restore left variables and secrets
		case base.SettingTypeEnvVar:
			changed, err := dropRestoreEnv(setting, resp)
			if err != nil {
				return nil, err
			}
			if changed {
				resp.Changed = append(resp.Changed, setting)
			}
		case base.SettingTypeSecret:
			if strings.HasPrefix(setting.Name, restoreSecretPrefix) {
				deleted[setting.ID] = true
				resp.Deleted = append(resp.Deleted, setting)
			}
		}
	}
	for _, setting := range req.Settings {
		if setting.Type != base.SettingTypeAppSettingMount {
			continue
		}
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if deleted[mount.Source.ID] {
			deleted[setting.ID] = true
			resp.Deleted = append(resp.Deleted, setting)
		}
	}

	locks := &req.Managed.Locks
	locks.EnvKeys = slices.DeleteFunc(locks.EnvKeys, isRestoreEnv)
	locks.Settings = slices.DeleteFunc(locks.Settings, func(id string) bool { return deleted[id] })
	if slices.Contains(locks.Mounts, restoreSourceDir) {
		resp.MountTargets = append(resp.MountTargets, restoreSourceDir)
		locks.Mounts = slices.DeleteFunc(locks.Mounts, func(target string) bool { return target == restoreSourceDir })
	}
	pg.Restore.ReadyAt = req.Now
	return resp, nil
}

// dropRestoreEnv takes the restore's variables out of an env-var setting, which
// it says whether it changed, adding their names to resp.
func dropRestoreEnv(setting *entity.Setting, resp *managedmodel.FinishRestoreResp) (bool, error) {
	vars, err := setting.AsEnvVars()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	kept := make([]*entity.EnvVar, 0, len(vars.Data))
	for _, v := range vars.Data {
		if isRestoreEnv(v.Key) {
			resp.EnvKeys = append(resp.EnvKeys, v.Key)
			continue
		}
		kept = append(kept, v)
	}
	if len(kept) == len(vars.Data) {
		return false, nil
	}
	vars.Data = kept
	if err = setting.SetData(vars); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	return true, nil
}

func isRestoreEnv(key string) bool {
	return strings.HasPrefix(key, restoreEnvPrefix)
}
