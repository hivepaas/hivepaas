package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging/mocks"
)

func TestInterfaces(t *testing.T) {
	var _ IDB = (*DB)(nil)
	var _ IDB = (*Tx)(nil)
}

func TestNewDB_ConnectionFailure(t *testing.T) {
	cfg := &config.Config{
		DB: config.DB{
			Host: "non-existent-host",
			Port: 5432,
		},
	}
	logger := &mocks.Logger{}

	db := NewDB(cfg, logger)

	assert.NotNil(t, db.DB)
	assert.NotEmpty(t, logger.Errors)
	assert.Contains(t, logger.Errors[0], "Failed to connect to database")
}
