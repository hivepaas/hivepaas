package base

const (
	EncryptionSaltPrefix = "hpsalt:"
)

type FileEncryptionFormat string

const (
	FileEncryptionNone      FileEncryptionFormat = ""
	FileEncryptionFormatAge FileEncryptionFormat = "age"
)

var (
	AllFileEncryptionFormats = []FileEncryptionFormat{FileEncryptionNone, FileEncryptionFormatAge}
)
