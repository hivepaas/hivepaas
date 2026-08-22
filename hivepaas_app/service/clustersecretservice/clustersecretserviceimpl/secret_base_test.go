package clustersecretserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func TestHasSecretChanges(t *testing.T) {
	s := &service{}

	t.Run("both nil", func(t *testing.T) {
		assert.False(t, s.HasSecretChanges(nil, nil))
	})

	t.Run("one nil", func(t *testing.T) {
		sec := &entity.Secret{Key: "DB_PASS"}
		assert.True(t, s.HasSecretChanges(sec, nil))
		assert.True(t, s.HasSecretChanges(nil, sec))
	})

	t.Run("identical basic fields and value", func(t *testing.T) {
		sec1 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123"), Base64: false}
		sec2 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123"), Base64: false}
		assert.False(t, s.HasSecretChanges(sec1, sec2))
	})

	t.Run("different value", func(t *testing.T) {
		sec1 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123")}
		sec2 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret456")}
		assert.True(t, s.HasSecretChanges(sec1, sec2))
	})

	t.Run("different key", func(t *testing.T) {
		sec1 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123")}
		sec2 := &entity.Secret{Key: "API_KEY", Value: entity.NewEncryptedField("secret123")}
		assert.True(t, s.HasSecretChanges(sec1, sec2))
	})

	t.Run("different SwarmRef presence", func(t *testing.T) {
		sec1 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{SecretID: "s1"}}
		sec2 := &entity.Secret{Key: "DB_PASS", Value: entity.NewEncryptedField("secret123")}
		assert.True(t, s.HasSecretChanges(sec1, sec2))
		assert.True(t, s.HasSecretChanges(sec2, sec1))
	})

	t.Run("different SecretID or SecretName", func(t *testing.T) {
		sec1 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID:   "s1",
				SecretName: "name1",
			},
		}
		sec2 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID:   "s1",
				SecretName: "name2",
			},
		}
		assert.True(t, s.HasSecretChanges(sec1, sec2))
	})

	t.Run("identical SwarmRef and File", func(t *testing.T) {
		sec1 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID:   "s1",
				SecretName: "name1",
				File: &entity.SwarmRefFileTarget{
					Name: "db.key",
					UID:  "0",
					GID:  "0",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		sec2 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID:   "s1",
				SecretName: "name1",
				File: &entity.SwarmRefFileTarget{
					Name: "db.key",
					UID:  "0",
					GID:  "0",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		assert.False(t, s.HasSecretChanges(sec1, sec2))
	})

	t.Run("different File target attributes", func(t *testing.T) {
		sec1 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID: "s1",
				File: &entity.SwarmRefFileTarget{
					Name: "db.key",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		sec2 := &entity.Secret{
			Key:   "DB_PASS",
			Value: entity.NewEncryptedField("secret123"),
			SwarmRef: &entity.SwarmSecretRef{
				SecretID: "s1",
				File: &entity.SwarmRefFileTarget{
					Name: "db.key",
					Mode: fileutil.FileMode(600),
				},
			},
		}
		assert.True(t, s.HasSecretChanges(sec1, sec2))
	})
}
