package entity

import (
	"time"
)

type DataMigration struct {
	ID        string    `bun:",pk" json:"id"`
	AppliedAt time.Time `json:"appliedAt"`
}

type DataMigrateable interface {
	Migrate() (hasChange bool, err error)
}
