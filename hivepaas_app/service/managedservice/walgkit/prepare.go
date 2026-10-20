package walgkit

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// OneArchive says whether an archive is exactly one of S3 and a volume.
func OneArchive(archive *managedmodel.Archive) bool {
	return archive != nil && (archive.S3 == nil) != (archive.Volume == nil)
}

// ArchiveEnv is where WAL-G finds an archive: an S3 prefix with its endpoint, or
// the directory a volume's is mounted at.
func ArchiveEnv(archive *managedmodel.Archive, volumeDir string) [][2]string {
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

// ArchiveSecrets are an archive's credentials, by the name of the file each is
// kept in: an S3 archive's keys, none for a volume's.
func ArchiveSecrets(archive *managedmodel.Archive) map[string]string {
	secrets := map[string]string{}
	if archive != nil && archive.S3 != nil {
		secrets[SecretAccessKey] = archive.S3.AccessKey
		secrets[SecretSecretKey] = archive.S3.SecretKey
	}
	return secrets
}

// CheckScripts refuses a document without every script the engine API runs.
func CheckScripts(p *Profile, doc *specmodel.AppDoc) error {
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
	for _, script := range p.Scripts {
		if path := ScriptDir + "/" + script; !present[path] {
			return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"engine %s API %d runs %s, which the template does not ship as a config file", p.Engine, p.API, path)
		}
	}
	return nil
}

// ScriptMentions says whether the template's script mentions text: what a
// script takes that an older template's did not.
func ScriptMentions(doc *specmodel.AppDoc, script, text string) bool {
	files, _ := doc.Settings["configFiles"].(map[string]any)
	for _, entry := range files {
		fields, _ := entry.(map[string]any)
		swarmRef, _ := fields["swarmRef"].(map[string]any)
		file, _ := swarmRef["file"].(map[string]any)
		if name, _ := file["name"].(string); name == ScriptDir+"/"+script {
			content, _ := fields["content"].(string)
			return strings.Contains(content, text)
		}
	}
	return false
}

// AddEnv adds variables to the document's.
func AddEnv(doc *specmodel.AppDoc, env [][2]string) {
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

// AddSecretFiles puts each secret in a file of dir, readable by owner alone,
// named prefix and the secret's name. Nothing refers to them from the
// environment, so they are never in it.
func AddSecretFiles(doc *specmodel.AppDoc, dir, prefix, owner string, secrets map[string]string) {
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
				"uid":  owner, "gid": owner, "mode": "400",
			}},
		}
	}
}

// MountArchiveVolume mounts a volume repository's directory at target.
func MountArchiveVolume(doc *specmodel.AppDoc, target string, volume *managedmodel.VolumeArchive, readOnly bool) {
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
