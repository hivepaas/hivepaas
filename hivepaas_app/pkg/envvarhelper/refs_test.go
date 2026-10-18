package envvarhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestRefersTo(t *testing.T) {
	for value, refers := range map[string]bool{
		"postgres://${db.HIVEPAAS_HOST}:5432": true,
		"${db.HIVEPAAS_DATABASE_URL}":         true,
		"${db-2.HIVEPAAS_HOST}":               false,
		"${dbx.HIVEPAAS_HOST}":                false,
		"${secrets.db}":                       false,
		"$db.HOST":                            false,
		"plain":                               false,
	} {
		assert.Equal(t, refers, RefersTo(&entity.EnvVar{Value: value}, "db"), value)
	}
	assert.False(t, RefersTo(&entity.EnvVar{Value: "${db.HOST}", IsLiteral: true}, "db"), "a literal names nothing")
}

// Only the references to the keys given are rewritten, and only in variables
// that are not literal; the variables changed are returned.
func TestRewriteRefs(t *testing.T) {
	vars := []*entity.EnvVar{
		{Key: "DATABASE_URL", Value: "${db.HIVEPAAS_DATABASE_URL}"},
		{Key: "DSN", Value: "host=${db.HIVEPAAS_HOST} port=${db.HIVEPAAS_PORT} cache=${cache.HIVEPAAS_HOST}"},
		{Key: "OTHER", Value: "${db-2.HIVEPAAS_HOST}"},
		{Key: "LITERAL", Value: "${db.HIVEPAAS_HOST}", IsLiteral: true},
		{Key: "SECRET", Value: "${secrets.db}"},
	}

	changed := RewriteRefs(vars, RefReplacer(map[string]string{"db": "db-restored"}))

	assert.Equal(t, []string{"DATABASE_URL", "DSN"}, keysOf(changed))
	assert.Equal(t, "${db-restored.HIVEPAAS_DATABASE_URL}", vars[0].Value)
	assert.Equal(t, "host=${db-restored.HIVEPAAS_HOST} port=${db-restored.HIVEPAAS_PORT} cache=${cache.HIVEPAAS_HOST}",
		vars[1].Value)
	assert.Equal(t, "${db-2.HIVEPAAS_HOST}", vars[2].Value)
	assert.Equal(t, "${db.HIVEPAAS_HOST}", vars[3].Value)
	assert.Equal(t, "${secrets.db}", vars[4].Value)
}

func keysOf(vars []*entity.EnvVar) []string {
	out := make([]string, 0, len(vars))
	for _, env := range vars {
		out = append(out, env.Key)
	}
	return out
}
