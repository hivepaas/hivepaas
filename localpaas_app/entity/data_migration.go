package entity

import (
	"time"
)

type DataMigration struct {
	ID        string `bun:",pk"`
	AppliedAt time.Time
}

type DataMigrateable interface {
	Migrate() (hasChange bool, err error)
}
