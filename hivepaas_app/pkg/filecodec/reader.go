// Package filecodec reads back the files a job saves its command's output to:
// compressed, then encrypted, as the suffixes of the file's name say.
package filecodec

import (
	"compress/gzip"
	"io"
	"strings"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	suffixAge  = ".age"
	suffixGzip = ".gz"
	suffixZstd = ".zst"
)

// FileLayers are what a file was wrapped in when it was saved.
type FileLayers struct {
	Compression base.FileCompressionFormat
	Encrypted   bool
}

// Layers are a file's, by its name: a job saving its output names the file
// with .gz or .zst for the compression, then .age for the encryption.
func Layers(name string) FileLayers {
	var layers FileLayers
	if rest, ok := strings.CutSuffix(name, suffixAge); ok {
		layers.Encrypted = true
		name = rest
	}
	switch {
	case strings.HasSuffix(name, suffixGzip):
		layers.Compression = base.FileCompressionFormatGzip
	case strings.HasSuffix(name, suffixZstd):
		layers.Compression = base.FileCompressionFormatZstd
	}
	return layers
}

// NewReader reads r, a file of the name given, as the command that made it
// printed it: decrypted with the passphrase, then decompressed.
func NewReader(r io.Reader, name, passphrase string) (io.Reader, error) {
	layers := Layers(name)
	if layers.Encrypted {
		if passphrase == "" {
			return nil, hperrors.NewMissing("Passphrase")
		}
		identity, err := age.NewScryptIdentity(passphrase)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if r, err = age.Decrypt(r, identity); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	switch layers.Compression {
	case base.FileCompressionFormatGzip:
		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return gz, nil
	case base.FileCompressionFormatZstd:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return zr.IOReadCloser(), nil
	case base.FileCompressionNone, base.FileCompressionFormatZip, base.FileCompressionFormatTar:
	}
	return r, nil
}
