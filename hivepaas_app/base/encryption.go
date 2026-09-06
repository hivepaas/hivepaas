package base

const (
	// EncryptionSaltPrefix marks a value derived-key encrypted with a per-value
	// salt. HashField still uses it; EncryptedField no longer does.
	EncryptionSaltPrefix = "hpsalt:"
	// EncryptionKeyPrefix marks a value sealed with the data encryption key.
	EncryptionKeyPrefix = "hpenc:"
)

// AllEncryptionPrefixes are the markers a stored value can carry. A value a user
// supplies must never start with one of them, or it would be read back as
// ciphertext instead of being encrypted.
var AllEncryptionPrefixes = []string{EncryptionSaltPrefix, EncryptionKeyPrefix}

// MaskedSecret is what an API response carries in place of a stored secret.
//
// It is deliberately one shared value. The placeholder used to be declared once
// per DTO package, in two different lengths ("********" and "****************"),
// so any check written against one of them silently missed every setting that
// used the other. It lives here rather than in basedto because entity has to
// recognize it too, and entity must not depend on the DTO layer.
const MaskedSecret = "********"

// IsMaskedSecret reports whether value is the placeholder rather than a real
// secret. A form that loads a setting and posts it back unchanged returns the
// placeholder, and storing it would destroy the secret it stands for.
func IsMaskedSecret(value string) bool {
	return value == MaskedSecret
}

type FileEncryptionFormat string

const (
	FileEncryptionNone      FileEncryptionFormat = ""
	FileEncryptionFormatAge FileEncryptionFormat = "age"
)

var (
	AllFileEncryptionFormats = []FileEncryptionFormat{FileEncryptionNone, FileEncryptionFormatAge}
)
