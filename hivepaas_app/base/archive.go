package base

type ArchiveType string

const (
	ArchiveTypeNone ArchiveType = ""
	ArchiveTypeTar  ArchiveType = "tar"
	ArchiveTypeGz   ArchiveType = "gz"
)

var (
	AllArchiveTypes = []ArchiveType{ArchiveTypeNone, ArchiveTypeTar, ArchiveTypeGz}
)
