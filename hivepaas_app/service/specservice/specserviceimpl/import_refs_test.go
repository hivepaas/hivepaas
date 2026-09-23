package specserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func routingBody(certs ...any) map[string]any {
	domains := make([]any, 0, len(certs))
	for i, cert := range certs {
		domains = append(domains, map[string]any{
			"enabled": true, "domain": "d" + string(rune('a'+i)) + ".example.com", "protocol": "http",
			"sslCert": map[string]any{"id": cert},
		})
	}
	return map[string]any{
		"port": 8080, "domains": domains,
		specmodel.SettingMetaKey: map[string]any{"status": "active", "version": 1},
	}
}

// A body is read as its type: the certificate held by path, the one held by an
// external block, and the one export could not resolve are each found, and the
// body the bundle holds is left as it was.
func TestSettingRefsReadsWhatTheTypeReports(t *testing.T) {
	external := map[string]any{externalRefKey: map[string]any{
		"type": "ssl-cert", "name": "localhost", "kind": "self-signed", "id": "cert_1",
	}}
	body := routingBody("global/sslCerts/localhost", external, "01GONE")

	refs, err := settingRefs(base.SettingTypeAppRouting, "projects/a/envs/dev/apps/web", "settings.routing", body)

	assert.NoError(t, err)
	if assert.Len(t, refs, 3) {
		byKind := map[refKind]bundleRef{}
		for _, ref := range refs {
			byKind[ref.kind] = ref
		}
		assert.Equal(t, "global/sslCerts/localhost", byKind[refPath].value)
		assert.Equal(t, &specmodel.ExternalRef{Type: "ssl-cert", Name: "localhost", Kind: "self-signed", ID: "cert_1"},
			byKind[refExternal].external)
		assert.Equal(t, "01GONE", byKind[refRawID].value)
		assert.Equal(t, "settings.routing", byKind[refPath].holder)
	}
	domains, _ := body["domains"].([]any)
	second, _ := domains[1].(map[string]any)
	assert.Equal(t, map[string]any{"id": external}, second["sslCert"], "the bundle's body is unchanged")
}

// Data written by a newer HivePaaS cannot be read here; SETTING_VERSION_NEWER
// already blocks it, so it has no references to follow.
func TestSettingRefsOfANewerVersionAreNone(t *testing.T) {
	body := routingBody("global/sslCerts/localhost")
	body[specmodel.SettingMetaKey] = map[string]any{"version": 999}

	refs, err := settingRefs(base.SettingTypeAppRouting, "global", "routing", body)

	assert.NoError(t, err)
	assert.Empty(t, refs)
}

func TestSettingRefsRefusesABodyThatIsNotItsType(t *testing.T) {
	_, err := settingRefs(base.SettingTypeAppRouting, "global", "routing", map[string]any{"port": "eighty"})

	assert.ErrorIs(t, err, hperrors.ErrSpecBundleInvalid)
}

func TestMountRefs(t *testing.T) {
	refs := mountRefs(&specmodel.Storage{Mounts: map[string]specmodel.Mount{
		"/data": {Type: mount.TypeVolume, Source: "projects/a/volumes/default",
			SourceApp: &specmodel.MountSourceApp{App: "db"}},
		"/shared": {Type: mount.TypeVolume, External: &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared"}},
	}})

	assert.Equal(t, []bundleRef{
		{kind: refPath, value: "projects/a/volumes/default", in: "mount", holder: "/data"},
		{kind: refSourceApp, value: "db", in: "mount", holder: "/data"},
		{kind: refExternal, value: "shared", in: "mount", holder: "/shared",
			external: &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared"}},
	}, refs)
}

func TestParseRefPath(t *testing.T) {
	cases := map[string]refTarget{
		"global/sslCerts/localhost": {node: "global", block: "sslCerts", key: "localhost",
			settingType: base.SettingTypeSSLCert, isCollectionSetting: true},
		"global/registryAuths/ghcr.io/acme": {node: "global", block: "registryAuths", key: "ghcr.io/acme",
			settingType: base.SettingTypeRegistryAuth, isCollectionSetting: true},
		"projects/a/volumes/default": {node: "projects/a/settings", project: "a", block: "volumes",
			key: "default", settingType: base.SettingTypeClusterVolume, isCollectionSetting: true},
		"projects/a/envs/dev/basicAuths/admin": {node: "projects/a/envs/dev/settings", project: "a", env: "dev",
			block: "basicAuths", key: "admin", settingType: base.SettingTypeBasicAuth, isCollectionSetting: true},
		"projects/a/envs/dev/apps/web/secrets/TOKEN": {node: "projects/a/envs/dev/apps/web", project: "a",
			env: "dev", app: "web", block: "secrets", key: "TOKEN", settingType: base.SettingTypeSecret,
			isCollectionSetting: true},
		"global/logging": {node: "global", block: "logging", settingType: base.SettingTypeLogging},
	}
	for path, want := range cases {
		got, ok := parseRefPath(path)
		assert.True(t, ok, path)
		assert.Equal(t, want, got, path)
	}

	for _, path := range []string{
		"global", "global/sslCerts", "global/nothing/x", "global/logging/x", "projects/a", "projects/a/envs/dev",
		"01J9ZKQ3YQ6V5W0000000000", "/etc/passwd",
	} {
		_, ok := parseRefPath(path)
		assert.False(t, ok, path)
	}
}
