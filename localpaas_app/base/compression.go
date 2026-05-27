package base

type FileCompressionFormat string

const (
	FileCompressionNone       FileCompressionFormat = ""
	FileCompressionFormatZstd FileCompressionFormat = "zstd"
	FileCompressionFormatGzip FileCompressionFormat = "gzip"
)

var (
	AllFileCompressionFormats = []FileCompressionFormat{FileCompressionNone, FileCompressionFormatZstd,
		FileCompressionFormatGzip}
)
