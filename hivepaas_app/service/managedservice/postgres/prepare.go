package postgres

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	// scriptDir is where the template's scripts are, beside the artifacts.
	scriptDir = "/opt/hivepaas/bin"
	// secretsDir is where the credentials are, one file each, readable by the
	// postgres user only. walg-run exports them for the one process it runs.
	secretsDir = "/run/secrets/walg" //nolint:gosec // G101: a directory
	// volumeArchiveDir is where a volume repository's directory is mounted.
	volumeArchiveDir = "/var/lib/hivepaas-walg"
	// postgresUID owns the secret files: the official Debian images' postgres
	// user, which archive_command runs as.
	postgresUID = "999"

	secretAccessKey = "WALG_AWS_ACCESS_KEY_ID"     //nolint:gosec // G101: a secret's name
	secretSecretKey = "WALG_AWS_SECRET_ACCESS_KEY" //nolint:gosec // G101: a secret's name
	secretSodiumKey = "WALG_LIBSODIUM_KEY"         //nolint:gosec // G101: a secret's name
)

// apiScripts are the scripts engine API 1 runs, which the template ships as
// config files in scriptDir. What each takes, prints and exits with is the
// contract; see the design §3.
var apiScripts = []string{
	"walg-run", "walg-fetch", "walg-nightly", "walg-health", "walg-restore-test", "walg-entrypoint", "walg-mark",
}

func (e *Engine) PrepareDoc(req *managedmodel.PrepareReq) error {
	doc := req.Doc
	if err := checkScripts(doc); err != nil {
		return err
	}
	archive := req.Archive
	if !oneArchive(archive) || (req.Restore != nil && !oneArchive(req.Restore.Archive)) {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("postgres: exactly one archive is required")
	}

	env := [][2]string{
		{"WALG_COMPRESSION_METHOD", "zstd"},
		{"WALG_PREVENT_WAL_OVERWRITE", "true"},
		// The key is 32 bytes, written as hex.
		{"WALG_LIBSODIUM_KEY_TRANSFORM", "hex"},
		{"HIVEPAAS_WALG_SECRETS_DIR", secretsDir},
	}
	env = append(env, archiveEnv(archive, volumeArchiveDir)...)
	if archive.Volume != nil {
		mountArchiveVolume(doc, volumeArchiveDir, archive.Volume, false)
	}
	secrets := e.ArchiveSecrets(archive)
	secrets[secretSodiumKey] = req.EncryptionKey
	addSecretFiles(doc, secretsDir, "", secrets)
	if req.Restore != nil {
		restoreEnv, err := e.prepareRestore(doc, req.Restore)
		if err != nil {
			return err
		}
		env = append(env, restoreEnv...)
	}
	addEnv(doc, env)
	req.OwnedEnv = make([]string, 0, len(env)+1)
	for _, pair := range env {
		req.OwnedEnv = append(req.OwnedEnv, pair[0])
	}
	// The image sets it, and the scripts read it: it is the engine's as well.
	req.OwnedEnv = append(req.OwnedEnv, "PGDATA")
	return nil
}

func oneArchive(archive *managedmodel.Archive) bool {
	return archive != nil && (archive.S3 == nil) != (archive.Volume == nil)
}

// archiveEnv is where WAL-G finds an archive: an S3 prefix with its endpoint, or
// the directory a volume's is mounted at.
func archiveEnv(archive *managedmodel.Archive, volumeDir string) [][2]string {
	if s3 := archive.S3; s3 != nil {
		return [][2]string{
			{"WALG_S3_PREFIX", "s3://" + s3.Bucket + "/" + s3.Prefix},
			{"AWS_ENDPOINT", s3.Endpoint},
			{"AWS_REGION", s3.Region},
			{"AWS_S3_FORCE_PATH_STYLE", strconv.FormatBool(s3.ForcePathStyle)},
		}
	}
	return [][2]string{{"WALG_FILE_PREFIX", volumeDir}}
}

func (e *Engine) ArchiveSecrets(archive *managedmodel.Archive) map[string]string {
	secrets := map[string]string{}
	if archive != nil && archive.S3 != nil {
		secrets[secretAccessKey] = archive.S3.AccessKey
		secrets[secretSecretKey] = archive.S3.SecretKey
	}
	return secrets
}

// checkScripts refuses a document without every script the engine API runs.
func checkScripts(doc *specmodel.AppDoc) error {
	present := map[string]bool{}
	files, _ := doc.Settings["configFiles"].(map[string]any)
	for _, entry := range files {
		fields, _ := entry.(map[string]any)
		swarmRef, _ := fields["swarmRef"].(map[string]any)
		file, _ := swarmRef["file"].(map[string]any)
		if name, ok := file["name"].(string); ok {
			present[name] = true
		}
	}
	for _, script := range apiScripts {
		if path := scriptDir + "/" + script; !present[path] {
			return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"engine postgres API %d runs %s, which the template does not ship as a config file", maxAPI, path)
		}
	}
	return nil
}

func addEnv(doc *specmodel.AppDoc, env [][2]string) {
	if doc.Settings == nil {
		doc.Settings = map[string]any{}
	}
	block, _ := doc.Settings["envVars"].(map[string]any)
	if block == nil {
		block = map[string]any{}
		doc.Settings["envVars"] = block
	}
	data, _ := block["data"].([]any)
	for _, pair := range env {
		data = append(data, map[string]any{"k": pair[0], "v": pair[1]})
	}
	block["data"] = data
}

// addSecretFiles puts each secret in a file of dir, readable by the postgres
// user alone, named prefix and the secret's name. Nothing refers to them from
// the environment, so they are never in it.
func addSecretFiles(doc *specmodel.AppDoc, dir, prefix string, secrets map[string]string) {
	if doc.Settings == nil {
		doc.Settings = map[string]any{}
	}
	block, _ := doc.Settings["secrets"].(map[string]any)
	if block == nil {
		block = map[string]any{}
		doc.Settings["secrets"] = block
	}
	for _, name := range slices.Sorted(maps.Keys(secrets)) {
		block[prefix+name] = map[string]any{
			"value": secrets[name],
			"swarmRef": map[string]any{"file": map[string]any{
				"name": fmt.Sprintf("%s/%s", dir, name),
				"uid":  postgresUID, "gid": postgresUID, "mode": "400",
			}},
		}
	}
}

func mountArchiveVolume(doc *specmodel.AppDoc, target string, volume *managedmodel.VolumeArchive, readOnly bool) {
	if doc.Deployment == nil {
		doc.Deployment = &specmodel.Deployment{}
	}
	if doc.Deployment.Storage == nil {
		doc.Deployment.Storage = &specmodel.Storage{}
	}
	if doc.Deployment.Storage.Mounts == nil {
		doc.Deployment.Storage.Mounts = map[string]specmodel.Mount{}
	}
	doc.Deployment.Storage.Mounts[target] = specmodel.Mount{
		Type: mount.TypeVolume, Source: volume.VolumeID, ReadOnly: readOnly,
		VolumeOptions: &specmodel.VolumeOptions{Subpath: volume.Subpath},
	}
}
