package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// postgresUID owns the secret files: the official Debian images' postgres
// user, which archive_command runs as.
const postgresUID = "999"

// apiScripts are the scripts engine API 1 runs, which the template ships as
// config files in walgkit.ScriptDir. What each takes, prints and exits with is
// the contract; see the design §3.
var apiScripts = []string{
	"walg-run", "walg-fetch", "walg-nightly", "walg-health", "walg-restore-test", "walg-entrypoint", "walg-mark",
}

func (e *Engine) PrepareDoc(req *managedmodel.PrepareReq) error {
	doc := req.Doc
	if err := walgkit.CheckScripts(profile, doc); err != nil {
		return hperrors.Wrap(err)
	}
	archive := req.Archive
	if !walgkit.OneArchive(archive) || (req.Restore != nil && !walgkit.OneArchive(req.Restore.Archive)) {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("postgres: exactly one archive is required")
	}

	env := [][2]string{
		{"WALG_COMPRESSION_METHOD", "zstd"},
		{"WALG_PREVENT_WAL_OVERWRITE", "true"},
		// The key is 32 bytes, written as hex.
		{"WALG_LIBSODIUM_KEY_TRANSFORM", "hex"},
		{walgkit.EnvSecretsDir, walgkit.SecretsDir},
	}
	env = append(env, walgkit.ArchiveEnv(archive, walgkit.VolumeArchiveDir)...)
	if archive.Volume != nil {
		walgkit.MountArchiveVolume(doc, walgkit.VolumeArchiveDir, archive.Volume, false)
	}
	secrets := e.ArchiveSecrets(archive)
	secrets[walgkit.SecretSodiumKey] = req.EncryptionKey
	walgkit.AddSecretFiles(doc, walgkit.SecretsDir, "", postgresUID, secrets)
	if req.Restore != nil {
		restoreEnv, err := e.prepareRestore(doc, req.Restore)
		if err != nil {
			return err
		}
		env = append(env, restoreEnv...)
	}
	walgkit.AddEnv(doc, env)
	req.OwnedEnv = make([]string, 0, len(env)+1)
	for _, pair := range env {
		req.OwnedEnv = append(req.OwnedEnv, pair[0])
	}
	// The image sets it, and the scripts read it: it is the engine's as well.
	req.OwnedEnv = append(req.OwnedEnv, "PGDATA")
	return nil
}

func (e *Engine) ArchiveSecrets(archive *managedmodel.Archive) map[string]string {
	return walgkit.ArchiveSecrets(archive)
}
