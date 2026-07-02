package config

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"

type Files struct {
	RequestMaxFile    int           `toml:"request_max_file" env:"HP_FILES_REQUEST_MAX_FILE" default:"10"`
	RequestMaxSize    unit.DataSize `toml:"request_max_size" env:"HP_FILES_REQUEST_MAX_SIZE" default:"10gb"`
	FileNameMaxLength int           `toml:"file_name_max_length" env:"HP_FILES_FILE_NAME_MAX_LENGTH" default:"100"`

	// Build Source Tarballs
	BuildSourceFileExts []string      `toml:"build_source_exts" env:"HP_FILES_BUILD_SOURCE_EXTS" default:"[.tar.zst, .tar.gz, .tgz, .tar.lz4]"` //nolint:lll
	BuildSourceMaxFile  int           `toml:"build_source_max_file" env:"HP_FILES_BUILD_SOURCE_MAX_FILE" default:"1"`
	BuildSourceMaxSize  unit.DataSize `toml:"build_source_max_size" env:"HP_FILES_BUILD_SOURCE_MAX_SIZE" default:"10gb"`
}
