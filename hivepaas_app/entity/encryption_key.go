package entity

import (
	"time"
)

// EncryptionKey is a data encryption key as stored, wrapped by the app secret.
//
// The raw key never touches the database: WrappedKey only opens with the app
// secret, which lives outside it. Whoever gets a database dump gets neither the
// values nor the key that would open them.
type EncryptionKey struct {
	ID         string `bun:",pk" json:"id"`
	WrappedKey string `json:"wrappedKey"`
	IsActive   bool   `json:"isActive"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:",default:current_timestamp" json:"updatedAt"`
}

// GetID implements IDEntity interface
func (k *EncryptionKey) GetID() string {
	return k.ID
}
