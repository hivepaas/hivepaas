package clustersecretserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func TestHasConfigChanges(t *testing.T) {
	s := &service{}

	t.Run("both nil", func(t *testing.T) {
		assert.False(t, s.HasConfigChanges(nil, nil))
	})

	t.Run("one nil", func(t *testing.T) {
		cfg := &entity.ConfigFile{Name: "app.conf"}
		assert.True(t, s.HasConfigChanges(cfg, nil))
		assert.True(t, s.HasConfigChanges(nil, cfg))
	})

	t.Run("identical basic fields", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{Name: "app.conf", Content: "foo=bar", Base64: false}
		cfg2 := &entity.ConfigFile{Name: "app.conf", Content: "foo=bar", Base64: false}
		assert.False(t, s.HasConfigChanges(cfg1, cfg2))
	})

	t.Run("different basic fields", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{Name: "app.conf", Content: "foo=bar"}
		cfg2 := &entity.ConfigFile{Name: "app.conf", Content: "foo=baz"}
		assert.True(t, s.HasConfigChanges(cfg1, cfg2))
	})

	t.Run("different SwarmRef presence", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{Name: "app.conf", SwarmRef: &entity.SwarmConfigRef{ConfigID: "123"}}
		cfg2 := &entity.ConfigFile{Name: "app.conf"}
		assert.True(t, s.HasConfigChanges(cfg1, cfg2))
		assert.True(t, s.HasConfigChanges(cfg2, cfg1))
	})

	t.Run("different ConfigID or ConfigName", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{Name: "app.conf", SwarmRef: &entity.SwarmConfigRef{ConfigID: "123", ConfigName: "cfg1"}}
		cfg2 := &entity.ConfigFile{Name: "app.conf", SwarmRef: &entity.SwarmConfigRef{ConfigID: "123", ConfigName: "cfg2"}}
		assert.True(t, s.HasConfigChanges(cfg1, cfg2))
	})

	t.Run("identical SwarmRef and File", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{
			Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{
				ConfigID:   "123",
				ConfigName: "cfg1",
				File: &entity.SwarmRefFileTarget{
					Name: "app.conf",
					UID:  "0",
					GID:  "0",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		cfg2 := &entity.ConfigFile{
			Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{
				ConfigID:   "123",
				ConfigName: "cfg1",
				File: &entity.SwarmRefFileTarget{
					Name: "app.conf",
					UID:  "0",
					GID:  "0",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		assert.False(t, s.HasConfigChanges(cfg1, cfg2))
	})

	t.Run("different File target attributes", func(t *testing.T) {
		cfg1 := &entity.ConfigFile{
			Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{
				ConfigID: "123",
				File: &entity.SwarmRefFileTarget{
					Name: "app.conf",
					Mode: fileutil.FileMode(444),
				},
			},
		}
		cfg2 := &entity.ConfigFile{
			Name: "app.conf",
			SwarmRef: &entity.SwarmConfigRef{
				ConfigID: "123",
				File: &entity.SwarmRefFileTarget{
					Name: "app.conf",
					Mode: fileutil.FileMode(600),
				},
			},
		}
		assert.True(t, s.HasConfigChanges(cfg1, cfg2))
	})
}
