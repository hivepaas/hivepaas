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
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "pg_dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "pg_dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_dump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading pg_restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "pg_restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "pg_restore (pipe)", setting.Name)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_restore")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mysqldump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mysqldump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mysqldump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mysqldump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mysql pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mysql.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mysql (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mysql")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mariadb-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mariadb-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mariadb-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mariadb-dump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mariadb pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mariadb.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mariadb (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mariadb")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mongodump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mongodump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mongodump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mongodump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading mongorestore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "mongorestore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "mongorestore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "mongorestore")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading redis-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "redis-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "redis-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "redis-cli")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading redis-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "redis-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "redis-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading clickhouse-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "clickhouse-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "clickhouse-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "clickhouse-client")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading clickhouse-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "clickhouse-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "clickhouse-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "clickhouse-client")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlite-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "sqlite-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "sqlite-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlite3")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlite-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "sqlite-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "sqlite-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlite3")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlserver-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "sqlserver-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "sqlserver-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlcmd")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading sqlserver-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "sqlserver-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "sqlserver-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sqlcmd")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading influx-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "influx-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "influx-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading influx-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "influx-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "influx-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading elasticsearch-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "elasticsearch-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "elasticsearch-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "elasticdump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading elasticsearch-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "elasticsearch-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "elasticsearch-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "elasticdump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading dolt-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "dolt-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "dolt-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading dolt-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "dolt-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "dolt-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "sh -c")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading neon-dump pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "neon-dump.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "neon-dump (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_dump")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	t.Run("success loading neon-restore pipe command template", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "neon-restore.pipe")
		assert.NoError(t, err)
		if assert.NotNil(t, setting) {
			assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
			assert.Equal(t, string(base.CommandTemplateBackup), setting.Kind)
			assert.Equal(t, "neon-restore (pipe)", setting.Name)
			assert.Equal(t, base.SettingStatusActive, setting.Status)

			cmdData, err := setting.AsCommandTemplate()
			assert.NoError(t, err)
			if assert.NotNil(t, cmdData) {
				assert.Contains(t, cmdData.Command, "pg_restore")
				assert.NotEmpty(t, cmdData.Link)
				assert.NotEmpty(t, cmdData.Desc)
				assert.NotEmpty(t, cmdData.ArgGroups)
			}
		}
	})

	diagnosticsTemplates := []struct {
		name        string
		expectedCmd string
	}{
		{"pg_stat_activity", "psql"},
		{"pg_table_size", "psql"},
		{"pg_ping", "psql"},
		{"mysql_processlist", "mysql"},
		{"mysql_table_size", "mysql"},
		{"mysql_ping", "mysqladmin"},
		{"mariadb_processlist", "mariadb"},
		{"mariadb_table_size", "mariadb"},
		{"mariadb_ping", "mariadb-admin"},
		{"redis_info", "redis-cli"},
		{"redis_bigkeys", "redis-cli"},
		{"redis_ping", "redis-cli"},
		{"mongo_server_status", "mongosh"},
		{"mongo_db_stats", "mongosh"},
		{"mongo_ping", "mongosh"},
		{"clickhouse_processes", "clickhouse-client"},
		{"elasticsearch_cluster_health", "curl"},
	}

	for _, tc := range diagnosticsTemplates {
		t.Run("success loading "+tc.name+" diagnostics command template", func(t *testing.T) {
			setting, err := s.GetCommand(ctx, string(base.CommandTemplateDiagnostics), tc.name)
			assert.NoError(t, err)
			if assert.NotNil(t, setting) {
				assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
				assert.Equal(t, string(base.CommandTemplateDiagnostics), setting.Kind)
				assert.Equal(t, tc.name, setting.Name)
				assert.Equal(t, base.SettingStatusActive, setting.Status)

				cmdData, err := setting.AsCommandTemplate()
				assert.NoError(t, err)
				if assert.NotNil(t, cmdData) {
					assert.Contains(t, cmdData.Command, tc.expectedCmd)
					assert.NotEmpty(t, cmdData.Link)
					assert.NotEmpty(t, cmdData.Desc)
					assert.NotEmpty(t, cmdData.ArgGroups)
				}
			}
		})
	}

	maintenanceTemplates := []struct {
		name        string
		expectedCmd string
	}{
		{"pg_vacuum_analyze", "psql"},
		{"pg_reindex_database", "psql"},
		{"mysql_optimize", "mysqlcheck"},
		{"mysql_analyze", "mysqlcheck"},
		{"mariadb_optimize", "mariadb-check"},
		{"mariadb_analyze", "mariadb-check"},
		{"redis_memory_purge", "redis-cli"},
		{"redis_bgsave", "redis-cli"},
		{"mongo_compact", "mongosh"},
		{"clickhouse_optimize_table", "clickhouse-client"},
		{"elasticsearch_forcemerge", "curl"},
		{"elasticsearch_clearcache", "curl"},
		{"docker_prune_all", "docker"},
		{"docker_logs_cleanup", "sh -c"},
	}

	for _, tc := range maintenanceTemplates {
		t.Run("success loading "+tc.name+" maintenance command template", func(t *testing.T) {
			setting, err := s.GetCommand(ctx, string(base.CommandTemplateMaintenance), tc.name)
			assert.NoError(t, err)
			if assert.NotNil(t, setting) {
				assert.Equal(t, base.SettingTypeCommandTemplate, setting.Type)
				assert.Equal(t, string(base.CommandTemplateMaintenance), setting.Kind)
				assert.Equal(t, tc.name, setting.Name)
				assert.Equal(t, base.SettingStatusActive, setting.Status)

				cmdData, err := setting.AsCommandTemplate()
				assert.NoError(t, err)
				if assert.NotNil(t, cmdData) {
					assert.Contains(t, cmdData.Command, tc.expectedCmd)
					assert.NotEmpty(t, cmdData.Link)
					assert.NotEmpty(t, cmdData.Desc)
				}
			}
		})
	}

	t.Run("nonexistent command template returns error", func(t *testing.T) {
		setting, err := s.GetCommand(ctx, string(base.CommandTemplateBackup), "nonexistent_cmd.pipe")
		assert.Error(t, err)
		assert.Nil(t, setting)
	})
}
