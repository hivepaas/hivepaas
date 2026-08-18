package base

type FileCompressionFormat string

const (
	FileCompressionNone       FileCompressionFormat = ""
	FileCompressionFormatZstd FileCompressionFormat = "zstd"
	FileCompressionFormatGzip FileCompressionFormat = "gzip"
	FileCompressionFormatZip  FileCompressionFormat = "zip"
	FileCompressionFormatTar  FileCompressionFormat = "tar"
)

var (
	AllFileCompressionFormats = []FileCompressionFormat{
		FileCompressionNone,
		FileCompressionFormatZstd,
		FileCompressionFormatGzip,
		FileCompressionFormatZip,
		FileCompressionFormatTar,
	}
)
