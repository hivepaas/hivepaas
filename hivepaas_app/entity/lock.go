package entity

import "time"

var (
	LockUpsertingConflictCols = []string{"id"}
	LockUpsertingUpdateCols   = []string{}
)

type Lock struct {
	ID string `bun:",pk" json:"id"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// GetID implements IDEntity interface
func (l *Lock) GetID() string {
	return l.ID
}
