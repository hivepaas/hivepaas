package walgkit

import (
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// What a restored app is restored with, and what walg-entrypoint reads it from,
// as engine API 1's contract (design §4, "The steps"):
//
//   - The source's archive is told as the app's own is, each variable's name
//     prefixed with RestoreEnvPrefix, and its keys are files of
//     RestoreSecretsDir, with their usual names. The scripts read the source
//     with those in place of the app's own; nothing writes to it.
//   - EnvRestoreBackup is the base backup to fetch: a name, or RestoreLatest.
//     Without it, EnvRestoreBefore is a time, and the backup fetched is the
//     latest that starts before it.
//   - Where it recovers to is the engine's (Profile.RestoreTarget).
//
// Every variable of the restore starts with RestoreEnvPrefix, and every secret
// with restoreSecretPrefix: that is how FinishRestore finds them.
const (
	RestoreEnvPrefix     = "HIVEPAAS_RESTORE_"
	EnvRestoreSecretsDir = RestoreEnvPrefix + "SECRETS_DIR"
	EnvRestoreBackup     = RestoreEnvPrefix + "TARGET_BACKUP"
	EnvRestoreBefore     = RestoreEnvPrefix + "TARGET_BEFORE"
	RestoreSecretPrefix  = "RESTORE_"
	RestoreSecretsDir    = SecretsDir + "/restore" //nolint:gosec // G101: a directory
	// RestoreSourceDir is where a volume archive the app is restored from is
	// mounted, read-only.
	RestoreSourceDir = "/var/lib/hivepaas-walg-source"
	// RestoreLatest is backup-fetch's name for the newest base backup.
	RestoreLatest = "LATEST"
)

// PrepareRestore adds to a restored app's document the source's archive and its
// keys, and returns the variables to add: those and the target's.
func PrepareRestore(p *Profile, doc *specmodel.AppDoc, source *managedmodel.RestoreSource) ([][2]string, error) {
	target, err := p.RestoreTarget(source)
	if err != nil {
		return nil, err
	}
	env := [][2]string{{EnvRestoreSecretsDir, RestoreSecretsDir}}
	for _, pair := range ArchiveEnv(source.Archive, RestoreSourceDir) {
		env = append(env, [2]string{RestoreEnvPrefix + pair[0], pair[1]})
	}
	if source.Archive.Volume != nil {
		MountArchiveVolume(doc, RestoreSourceDir, source.Archive.Volume, true)
	}
	secrets := ArchiveSecrets(source.Archive)
	secrets[SecretSodiumKey] = source.EncryptionKey
	AddSecretFiles(doc, RestoreSecretsDir, RestoreSecretPrefix, p.FileOwner, secrets)
	return append(env, target...), nil
}

// RecoveryCheck asks walg-entrypoint: up and serving, still recovering, or
// failed and why.
func RecoveryCheck() managedmodel.RecoveryCheck {
	return managedmodel.RecoveryCheck{Command: ScriptDir + "/walg-entrypoint status"}
}

// FirstBackupJob is the nightly job: its run takes the restored app's first base
// backup, into its own archive.
func FirstBackupJob(managed *entity.AppManagedSettings) string {
	if managed == nil || managed.Database == nil {
		return ""
	}
	return managed.Database.Jobs.Nightly.ID
}

// FinishRestore takes off a restored app what it was restored with: the
// restore's variables, its secrets and their mounts, and the source archive's
// mount; and marks the restore ready.
func FinishRestore(p *Profile, req *managedmodel.FinishRestoreReq) (*managedmodel.FinishRestoreResp, error) {
	db := req.Managed.Database
	if db == nil || db.Restore == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("%s: not a restored database", p.Engine)
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
			if strings.HasPrefix(setting.Name, RestoreSecretPrefix) {
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
	if slices.Contains(locks.Mounts, RestoreSourceDir) {
		resp.MountTargets = append(resp.MountTargets, RestoreSourceDir)
		locks.Mounts = slices.DeleteFunc(locks.Mounts, func(target string) bool { return target == RestoreSourceDir })
	}
	db.Restore.ReadyAt = req.Now
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
	return strings.HasPrefix(key, RestoreEnvPrefix)
}
