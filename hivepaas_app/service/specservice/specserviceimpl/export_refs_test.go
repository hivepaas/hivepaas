package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// A project exported alone does not hold the global certificate its app uses.
// The reference becomes an external one: the id finds the certificate again on
// the installation that exported it, the type, name and kind anywhere else.
func TestExportWritesAReferenceOutsideTheExportAsExternal(t *testing.T) {
	path, _ := runExportAt(t, entity.NewObjectScopeProject("p1"), specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": map[string]any{"external": map[string]any{
		"type": "ssl-cert", "name": "localhost", "kind": "self-signed", "id": "cert_1",
	}}}, domain["sslCert"])
}

// The whole installation holds the certificate, so the same reference is a path.
func TestExportWritesAReferenceInsideTheExportAsAPath(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": "global/sslCerts/localhost"}, domain["sslCert"])
}
