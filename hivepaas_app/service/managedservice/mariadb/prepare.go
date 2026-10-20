package mariadb

import (
	"strconv"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// envArchiveTimeout is walg-archive's period, in seconds: how often the archive
// loop closes the binary log, if it has events, and pushes it. walg-health
// reads it too.
const envArchiveTimeout = "HIVEPAAS_ARCHIVE_TIMEOUT"

// localhostAccount is what makes the image create mysql@localhost, which the
// unix socket identifies, with every privilege: the account the scripts reach
// the server as, run as the owner of the data files. No password is kept.
var localhostAccount = [][2]string{
	{"MARIADB_MYSQL_LOCALHOST_USER", "1"},
	{"MARIADB_MYSQL_LOCALHOST_GRANTS", "ALL"},
}

func (e *Engine) PrepareDoc(req *managedmodel.PrepareReq) error {
	doc := req.Doc
	if err := walgkit.CheckScripts(profile, doc); err != nil {
		return hperrors.Wrap(err)
	}
	archive := req.Archive
	if !walgkit.OneArchive(archive) || (req.Restore != nil && !walgkit.OneArchive(req.Restore.Archive)) {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("mariadb: exactly one archive is required")
	}
	options := walgkit.WithDefaults(req.Options)

	env := [][2]string{
		{"WALG_COMPRESSION_METHOD", "zstd"},
		// The key is 32 bytes, written as hex.
		{"WALG_LIBSODIUM_KEY_TRANSFORM", "hex"},
		{walgkit.EnvSecretsDir, walgkit.SecretsDir},
		{envArchiveTimeout, strconv.Itoa(int(options.ArchiveTimeout / time.Second))},
	}
	env = append(env, walgkit.ArchiveEnv(archive, walgkit.VolumeArchiveDir)...)
	if archive.Volume != nil {
		walgkit.MountArchiveVolume(doc, walgkit.VolumeArchiveDir, archive.Volume, false)
	}
	secrets := e.ArchiveSecrets(archive)
	secrets[walgkit.SecretSodiumKey] = req.EncryptionKey
	walgkit.AddSecretFiles(doc, walgkit.SecretsDir, "", mysqlUID, secrets)
	if req.Restore != nil {
		if err := checkTarget(doc, req.Restore); err != nil {
			return hperrors.Wrap(err)
		}
		restoreEnv, err := walgkit.PrepareRestore(profile, doc, req.Restore)
		if err != nil {
			return hperrors.Wrap(err)
		}
		env = append(env, restoreEnv...)
	}
	// The template sets the account's variables; one that does not gets them.
	have := docEnvKeys(doc)
	for _, pair := range localhostAccount {
		if !have[pair[0]] {
			env = append(env, pair)
		}
	}
	walgkit.AddEnv(doc, env)
	req.OwnedEnv = make([]string, 0, len(env)+len(localhostAccount))
	for _, pair := range env {
		req.OwnedEnv = append(req.OwnedEnv, pair[0])
	}
	for _, pair := range localhostAccount {
		if have[pair[0]] {
			req.OwnedEnv = append(req.OwnedEnv, pair[0])
		}
	}
	return nil
}

func (e *Engine) ArchiveSecrets(archive *managedmodel.Archive) map[string]string {
	return walgkit.ArchiveSecrets(archive)
}

// docEnvKeys are the variables the document sets.
func docEnvKeys(doc *specmodel.AppDoc) map[string]bool {
	keys := map[string]bool{}
	block, _ := doc.Settings["envVars"].(map[string]any)
	data, _ := block["data"].([]any)
	for _, entry := range data {
		if pair, ok := entry.(map[string]any); ok {
			if key, ok := pair["k"].(string); ok {
				keys[key] = true
			}
		}
	}
	return keys
}
