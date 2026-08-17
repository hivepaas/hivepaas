package commandserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestGetCommand(t *testing.T) {
	s := &service{}
	ctx := context.Background()

	t.Run("success loading pg_dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "pg_dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "pg_dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_dump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading pg_restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "pg_restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "pg_restore (pipe)", setting.Name)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_restore")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mysqldump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mysqldump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mysqldump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mysqldump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mysql pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mysql", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mysql (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mysql")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mariadb-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mariadb-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mariadb-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mariadb-dump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mariadb pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mariadb", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mariadb (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mariadb")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mongodump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mongodump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mongodump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mongodump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mongorestore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "mongorestore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "mongorestore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mongorestore")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading redis-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "redis-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "redis-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "redis-cli")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading redis-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "redis-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "redis-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading clickhouse-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "clickhouse-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "clickhouse-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "clickhouse-client")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading clickhouse-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "clickhouse-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "clickhouse-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "clickhouse-client")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlite-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "sqlite-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "sqlite-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlite3")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlite-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "sqlite-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "sqlite-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlite3")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlserver-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "sqlserver-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "sqlserver-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlcmd")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlserver-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "sqlserver-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "sqlserver-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlcmd")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading influx-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "influx-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "influx-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "influx")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading influx-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "influx-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "influx-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "influx")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading elasticsearch-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "elasticsearch-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "elasticsearch-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "elasticdump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading elasticsearch-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "elasticsearch-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "elasticsearch-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "elasticdump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading dolt-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "dolt-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "dolt-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "dolt")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading dolt-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "dolt-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "dolt-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "dolt")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading neon-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "neon-dump", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "neon-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_dump")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading neon-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "neon-restore", "pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateDatabase), setting.Kind)
			assert.Equal(t, "neon-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_restore")
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("nonexistent command template returns error", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, "nonexistent_cmd", "pipe")
		assert.Error(t, err)
		assert.Nil(t, setting)
	})
}
