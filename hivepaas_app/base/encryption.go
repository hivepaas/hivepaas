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

type FileEncryptionFormat string

const (
	FileEncryptionNone      FileEncryptionFormat = ""
	FileEncryptionFormatAge FileEncryptionFormat = "age"
)

var (
	AllFileEncryptionFormats = []FileEncryptionFormat{FileEncryptionNone, FileEncryptionFormatAge}
)
