package composeserviceimpl

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/consts"
	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/services/docker"
)

// The loader serves the request's files and nothing else, written to the
// scratch directory once asked for; a relative path is the naming file's.
func TestSandboxLoadsOnlyTheRequestsFiles(t *testing.T) {
	dir := t.TempDir()
	sb := newSandbox(dir, map[string][]byte{
		"base.yaml": []byte("x"), "sub/inc.yaml": []byte("y"), "sub/b.yaml": []byte("z"), "sub/.env": []byte("A=1"),
	})

	got, err := sb.Load(context.Background(), "base.yaml")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "base.yaml"), got)
	assert.FileExists(t, got, "written once asked for")
	assert.NoFileExists(t, filepath.Join(dir, "sub", "b.yaml"), "not before")
	got, err = sb.Load(context.Background(), filepath.Join(dir, "base.yaml"))
	assert.NoError(t, err, "an absolute path inside it")
	assert.Equal(t, filepath.Join(dir, "base.yaml"), got)

	inSub := context.WithValue(context.Background(), consts.ComposeFileKey{}, filepath.Join(dir, "sub", "inc.yaml"))
	got, err = sb.Load(inSub, "b.yaml")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "sub", "b.yaml"), got, "beside the file naming it")
	assert.FileExists(t, filepath.Join(dir, "sub", ".env"), "the .env beside it, which compose-go reads")
	assert.Equal(t, filepath.Join(dir, "sub"), sb.Dir("b.yaml"), "asked with the path as written")
	assert.Equal(t, filepath.Join(dir, "sub"), sb.Dir(got))
	assert.Equal(t, dir, sb.Dir("/etc/hostname"), "never a directory outside")
	_, err = sb.Load(inSub, "../../x")
	assert.ErrorIs(t, err, hperrors.ErrComposeFilePath)

	for _, p := range []string{"/etc/hostname", "../x", filepath.Join(dir, "..", "x"), "https://example.com/c.yaml"} {
		_, err = sb.Load(context.Background(), p)
		assert.ErrorIs(t, err, hperrors.ErrComposeFilePath, p)
	}
	_, err = sb.Load(context.Background(), "other.yaml")
	assert.ErrorIs(t, err, hperrors.ErrComposeInvalid, "not among the files given")
	assert.Equal(t, []string{"other.yaml"}, sb.missingFiles())
	assert.Equal(t, []string{"base.yaml", "sub/b.yaml"}, sb.loadedFiles())
}

const sharedCompose = `
services:
  api:
    image: app:1
    volumes: [storage:/app/storage, ./uploads:/app/uploads]
  worker:
    image: app:1
    depends_on: [api]
    volumes: [storage:/app/storage:ro, ./uploads:/app/uploads]
volumes:
  storage:
`

// A volume - or a directory of the compose file's - two services mount is
// the first writer's directory, which the other mounts.
func TestConvertSharesAVolumeThroughItsOwner(t *testing.T) {
	resp := convert(t, convertReq(sharedCompose))
	api, worker := appOf(t, resp, "api"), appOf(t, resp, "worker")

	assert.Nil(t, api.Deployment.Storage.Mounts["/app/storage"].SourceApp)
	assert.Equal(t, &specmodel.MountSourceApp{App: "api"}, worker.Deployment.Storage.Mounts["/app/storage"].SourceApp,
		"read only")
	assert.Equal(t, "storage", worker.Deployment.Storage.Mounts["/app/storage"].VolumeOptions.Subpath)

	uploads := worker.Deployment.Storage.Mounts["/app/uploads"]
	assert.Equal(t, &specmodel.MountSourceApp{App: "api", Write: true}, uploads.SourceApp)
	assert.Equal(t, "uploads", uploads.VolumeOptions.Subpath)
	assert.Contains(t, codes(resp.Issues[envPath+"/apps/api"]), composeservice.CodeDirectoryEmpty)
}

const filesCompose = `
services:
  web:
    image: nginx:1.27
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs/site.pem:/etc/ssl/site.pem
      - /var/run/docker.sock:/var/run/docker.sock
      - /srv/data:/data
    tmpfs: [/run]
    secrets:
      - db_password
      - source: api_key
        target: api.key
        mode: 0400
    configs:
      - source: app_config
        target: /etc/app.json
secrets:
  db_password:
    file: ./secrets/db_password.txt
  api_key:
    environment: API_KEY
configs:
  app_config:
    content: '{"debug": false}'
`

// The compose file's files become env settings mounted where compose mounts
// them; one missing is created empty, pending, and asked for; the host's
// directories are left out for a caller who may not bind them.
func TestConvertMountsFilesFromEnvSettings(t *testing.T) {
	req := convertReq(filesCompose)
	req.DotEnv = "API_KEY=k3y\n"
	req.Files = map[string][]byte{"nginx.conf": []byte("events {}\n"), "secrets/db_password.txt": []byte("pw")}
	resp := convert(t, req)

	env := resp.Bundle.Envs["blog"]["prod"]
	configs, _ := env.Settings["configFiles"].(map[string]any)
	secrets, _ := env.Settings["secrets"].(map[string]any)
	assert.Equal(t, "events {}\n", configs["nginx.conf"].(map[string]any)["content"])
	assert.Equal(t, "pending", configs["site.pem"].(map[string]any)["setting"].(map[string]any)["status"])
	assert.Equal(t, `{"debug": false}`, configs["app_config"].(map[string]any)["content"])
	assert.Equal(t, "pw", secrets["db_password"].(map[string]any)["value"])
	assert.Equal(t, "k3y", secrets["api_key"].(map[string]any)["value"])
	assert.NotContains(t, configs["nginx.conf"], fileSourceKey, "what marks a setting while it is made is gone")

	web := appOf(t, resp, "web")
	mounts, _ := web.Settings["settingMounts"].(map[string]any)
	apiKey, _ := mounts["secret-api-key"].(map[string]any)
	assert.Equal(t, map[string]any{"id": envPath + "/secrets/api_key"}, apiKey["source"])
	assert.Equal(t, []any{map[string]any{"part": "value", "path": "/run/secrets/api.key", "mode": "0400"}},
		apiKey["files"])
	assert.Contains(t, mounts, "nginx-conf")
	assert.Contains(t, mounts, "config-app-config")
	for key := range mounts {
		assert.True(t, settingmountservice.ValidEntryKey(key), "an entry with key %q is never mounted", key)
	}

	storage := web.Deployment.Storage
	assert.Empty(t, storage.Mounts)
	assert.Equal(t, mount.TypeTmpfs, storage.DockerMounts["/run"].Type)
	assert.NotContains(t, storage.DockerMounts, "/data", "a host's directory, for a caller who may not")
	assert.NotContains(t, storage.DockerMounts, "/var/run/docker.sock")

	webIssues := codes(resp.Issues[envPath+"/apps/web"])
	assert.Contains(t, webIssues, composeservice.CodeMountDropped)
	assert.Contains(t, webIssues, composeservice.CodeFileMissing, "certs/site.pem, mounted empty")
	if assert.Len(t, resp.Needs, 3) {
		assert.Equal(t, "certs/site.pem", resp.Needs[0].Path)
		assert.False(t, resp.Needs[0].Given)
	}

	req.MayBindHost = true
	resp = convert(t, req)
	assert.Equal(t, mount.TypeBind, appOf(t, resp, "web").Deployment.Storage.DockerMounts["/data"].Type)
}

// A service only built is not created, unless the review gives it an image.
func TestConvertSkipsAServiceItCannotBuild(t *testing.T) {
	compose := "services:\n  app:\n    build: .\n  db: {image: postgres:17}\n"
	resp := convert(t, convertReq(compose))
	assert.Equal(t, []string{composeservice.CodeNoImage}, codes(resp.Issues[envPath+"/apps/app"]))
	assert.Equal(t, specmodel.SeveritySkipped, resp.Issues[envPath+"/apps/app"][0].Severity)

	req := convertReq(compose)
	req.Services = map[string]*composeservice.ServiceReq{"app": {Image: "ghcr.io/me/app:1"}}
	resp = convert(t, req)
	assert.Equal(t, "ghcr.io/me/app:1", appOf(t, resp, "app").Deployment.Source["imageSource"].(map[string]any)["image"])
	assert.Contains(t, codes(resp.Issues[envPath+"/apps/app"]), composeservice.CodeBuildIgnored)
}

// What Swarm has no field for is named, and left out.
func TestConvertNamesWhatItLeavesOut(t *testing.T) {
	compose := "services:\n  a:\n    image: x:1\n    privileged: true\n    network_mode: host\n" +
		"    cap_add: [NET_ADMIN]\n"
	resp := convert(t, convertReq(compose))
	issues := resp.Issues[envPath+"/apps/a"]
	assert.Contains(t, codes(issues), composeservice.CodeNotSupported)
	for _, issue := range issues {
		if issue.Code == composeservice.CodeNotSupported {
			assert.ElementsMatch(t, []string{"privileged", "network_mode"}, issue.Detail["fields"])
		}
	}
	assert.Contains(t, codes(issues), composeservice.CodeCapabilityDropped, "the caller may not grant them")
	assert.Nil(t, appOf(t, resp, "a").Deployment.Resources)

	req := convertReq(compose)
	req.MayWriteCluster = true
	assert.Equal(t, []string{"NET_ADMIN"}, appOf(t, convert(t, req), "a").Deployment.Resources.Capabilities.CapabilityAdd)
}

// Two services whose names make one app key, or that wait on each other,
// cannot be created as they are.
func TestConvertBlocksWhatCannotBeCreated(t *testing.T) {
	resp := convert(t, convertReq("services:\n  my_db: {image: x:1}\n  my-db: {image: x:1}\n"))
	assert.Equal(t, []string{composeservice.CodeKeyConflict}, codes(resp.Issues[envPath]))

	_, err := New().Convert(context.Background(),
		convertReq("services:\n  a: {image: x:1, depends_on: [b]}\n  b: {image: x:1, depends_on: [a]}\n"))
	assert.ErrorIs(t, err, hperrors.ErrComposeInvalid, "compose-go finds the cycle, and says where")
}

// A name the app key changes is still how the others reach it.
func TestConvertKeepsTheNamesAServiceIsReachedBy(t *testing.T) {
	resp := convert(t, convertReq("services:\n  My_DB:\n    image: x:1\n    container_name: legacy-db\n"+
		"  app:\n    image: x:1\n    links: ['My_DB:database']\n"))
	db := appOf(t, resp, "my-db")
	assert.Equal(t, []string{"my-db", "My_DB", "legacy-db", "database"}, db.Deployment.Networks.Attachments[0].Aliases)
}

// A migration another waits on to complete runs to completion.
func TestConvertRunsAOneOffStepAsAJob(t *testing.T) {
	resp := convert(t, convertReq(`
services:
  migrate: {image: app:1, command: migrate, restart: "no"}
  app:
    image: app:1
    depends_on: {migrate: {condition: service_completed_successfully}}
`))
	modeSpec := appOf(t, resp, "migrate").Deployment.Service.ModeSpec
	assert.Equal(t, docker.ServiceModeReplicatedJob, modeSpec.Mode)
	assert.Equal(t, docker.ServiceModeReplicated, appOf(t, resp, "app").Deployment.Service.ModeSpec.Mode)
	assert.Equal(t, "migrate", appOf(t, resp, "migrate").Deployment.Source["command"])
}

// extends, in the file or from a file given, and profiles, as compose reads
// them.
func TestConvertReadsExtendsAndProfiles(t *testing.T) {
	req := convertReq(`
services:
  base: {image: app:1, environment: {A: "1"}}
  web: {extends: base, environment: {B: "2"}}
  worker: {extends: {file: common.yaml, service: worker}}
  debug: {image: busybox:1, profiles: [debug]}
`)
	req.Files = map[string][]byte{"common.yaml": []byte("services:\n  worker: {image: app:1, command: work}\n")}
	resp := convert(t, req)
	assert.Equal(t, "1", envVar(appOf(t, resp, "web"), "A")["v"])
	assert.Equal(t, "2", envVar(appOf(t, resp, "web"), "B")["v"])
	assert.Equal(t, "work", appOf(t, resp, "worker").Deployment.Source["command"])
	assert.NotContains(t, resp.Bundle.Envs["blog"]["prod"].Apps, "debug", "its profile was not asked for")
	assert.Equal(t, []string{"debug"}, resp.Profiles)

	req.Profiles = []string{"debug"}
	assert.Contains(t, convert(t, req).Bundle.Envs["blog"]["prod"].Apps, "debug")
}

// Each mount is an entry of its own, keyed as an entry key has to be - one
// that is not is never mounted - and numbered when two would share one.
func TestEntryKeysAreValidAndUnique(t *testing.T) {
	mounts := map[string]any{}
	hints := []string{"nginx.conf", "nginx.conf", "secret-DB_PASSWORD", "___", "a-very-long-name-of-a-config-file.yaml"}
	for _, hint := range hints {
		addMount(mounts, hint, "source", map[string]any{})
	}
	keys := make([]string, 0, len(mounts))
	for key := range mounts {
		keys = append(keys, key)
		assert.True(t, settingmountservice.ValidEntryKey(key), key)
	}
	assert.ElementsMatch(t, []string{"nginx-conf", "nginx-conf-2", "secret-db-password", "file",
		"a-very-long-name-of"}, keys)
}

// A directory the compose file mounts is the app's own on the project's
// volume, as ever; the files given under it are mounted in it, read only, and
// what the app writes beside them is kept.
func TestConvertMountsTheFilesGivenUnderADirectory(t *testing.T) {
	req := convertReq(`
services:
  web:
    image: nginx:1.27
    volumes: ["./nginx/conf.d:/etc/nginx/conf.d", "./data:/data"]
`)
	req.Files = map[string][]byte{
		"nginx/conf.d/default.conf": []byte("server {}"), "nginx/conf.d/sub/x.conf": []byte("x"),
	}
	resp := convert(t, req)

	web := appOf(t, resp, "web")
	if assert.Contains(t, web.Deployment.Storage.Mounts, "/etc/nginx/conf.d") {
		assert.Equal(t, "nginx-conf.d", web.Deployment.Storage.Mounts["/etc/nginx/conf.d"].VolumeOptions.Subpath)
	}
	mounts, _ := web.Settings["settingMounts"].(map[string]any)
	var paths []any
	for key, body := range mounts {
		assert.True(t, settingmountservice.ValidEntryKey(key), key)
		files, _ := body.(map[string]any)["files"].([]any)
		for _, file := range files {
			paths = append(paths, file.(map[string]any)["path"])
		}
	}
	assert.ElementsMatch(t, []any{"/etc/nginx/conf.d/default.conf", "/etc/nginx/conf.d/sub/x.conf"}, paths)
	configs, _ := resp.Bundle.Envs["blog"]["prod"].Settings["configFiles"].(map[string]any)
	assert.Contains(t, configs, "default.conf")

	assert.ElementsMatch(t, []*composeservice.FileNeed{
		{Path: "nginx/conf.d", As: composeservice.NeedDirectory, By: []string{"web"}, Given: true},
		{Path: "data", As: composeservice.NeedDirectory, By: []string{"web"}},
	}, resp.Needs)
	issues := codes(resp.Issues[envPath+"/apps/web"])
	assert.Contains(t, issues, composeservice.CodeDirectoryFiles)
	assert.Contains(t, issues, composeservice.CodeDirectoryEmpty, "the one with no file given")
	for _, volume := range resp.Services[0].Volumes {
		if volume.Target == "/etc/nginx/conf.d" {
			assert.Equal(t, 2, volume.Files)
		}
	}
}

// A directory mounted read only holds nothing the app writes: the files given
// are the directory, with no volume under them - none could be mounted in a
// read-only one. Read from another app that writes it, it is that app's, and
// the files are not mounted.
func TestConvertMountsAReadOnlyDirectoryAsItsFiles(t *testing.T) {
	req := convertReq(`
services:
  web:
    image: nginx:1.27
    volumes: ["./conf.d:/etc/nginx/conf.d:ro"]
  certbot:
    image: certbot/certbot:v3
    volumes: ["./certs:/etc/letsencrypt"]
  proxy:
    image: nginx:1.27
    depends_on: [certbot]
    volumes: ["./certs:/etc/letsencrypt:ro"]
`)
	req.Files = map[string][]byte{"conf.d/default.conf": []byte("server {}"), "certs/options.conf": []byte("x")}
	resp := convert(t, req)

	web := appOf(t, resp, "web")
	assert.Nil(t, web.Deployment.Storage, "no volume")
	mounts, _ := web.Settings["settingMounts"].(map[string]any)
	assert.Len(t, mounts, 1)
	for _, view := range resp.Services {
		if view.Name == "web" {
			assert.Equal(t, composeservice.VolumeKindFiles, view.Volumes[0].Kind)
			assert.Equal(t, 1, view.Volumes[0].Files)
		}
	}

	proxy := appOf(t, resp, "proxy")
	if assert.NotNil(t, proxy.Deployment.Storage) {
		assert.NotNil(t, proxy.Deployment.Storage.Mounts["/etc/letsencrypt"].SourceApp, "certbot's directory")
	}
	assert.Nil(t, proxy.Settings["settingMounts"], "not in a read-only volume")
	certbotMounts, _ := appOf(t, resp, "certbot").Settings["settingMounts"].(map[string]any)
	assert.Len(t, certbotMounts, 1, "in the writer's own")
}

// A file mounted in a read-only volume would keep the app from starting -
// Docker has no place to make for it there - so the volume is writable.
func TestConvertMakesWritableAReadOnlyVolumeAFileIsMountedIn(t *testing.T) {
	req := convertReq(`
services:
  web:
    image: nginx:1.27
    volumes: ["site:/usr/share/nginx/html:ro", "./index.html:/usr/share/nginx/html/index.html", "logs:/logs:ro"]
volumes:
  site:
  logs:
`)
	req.Files = map[string][]byte{"index.html": []byte("<h1>hi</h1>")}
	resp := convert(t, req)

	mounts := appOf(t, resp, "web").Deployment.Storage.Mounts
	assert.False(t, mounts["/usr/share/nginx/html"].ReadOnly, "a file is mounted in it")
	assert.True(t, mounts["/logs"].ReadOnly, "none is")
	assert.Contains(t, codes(resp.Issues[envPath+"/apps/web"]), composeservice.CodeMountWritable)
}
