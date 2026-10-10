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

	// maxScryptWorkFactor is the scrypt work a passphrase is read with at most:
	// age's own, which a job's encryption uses. A file is anyone's to upload, and
	// the work its header names is memory and time the server spends.
	maxScryptWorkFactor = 18
	// maxZstdWindow is the window a zstd file is read with at most: what zstd
	// uses with --long, well above its levels' own.
	maxZstdWindow = 128 << 20
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
		identity.SetMaxWorkFactor(maxScryptWorkFactor)
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
		// Decoded on the reader's goroutine: one of its own would wait on a reader
		// left unread, and nothing closes it.
		zr, err := zstd.NewReader(r, zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxWindow(maxZstdWindow))
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return zr.IOReadCloser(), nil
	case base.FileCompressionNone, base.FileCompressionFormatZip, base.FileCompressionFormatTar:
	}
	return r, nil
}
