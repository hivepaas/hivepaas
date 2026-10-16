package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestRecordLicenseSeenAtOnlyRaisesTheTime(t *testing.T) {
	repo := &systemStatusRepo{}
	db := bun.NewDB(nil, pgdialect.New())
	seen := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	sql := repo.recordLicenseSeenAtQuery(db, seen).String()

	assert.Equal(t, `UPDATE "system_statuses" AS "system_status" `+
		`SET license_seen_at = GREATEST(license_seen_at, '2026-10-16 12:00:00+00:00') `+
		`WHERE (system_status.id = 1) RETURNING license_seen_at`, sql)
	assert.NotContains(t, sql, "update_ver", "a time seen is not a change anything acts on")
}
