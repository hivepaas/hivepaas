package filecodec

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"filippo.io/age"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const passphrase = "a passphrase of the job's"

// written is what a job saving its output writes for name: compressed as the
// name's .gz or .zst says, then encrypted as its .age says.
func written(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	var w io.WriteCloser = nopCloser{&out}
	closers := []io.Closer{}
	if layers := Layers(name); layers.Encrypted {
		recipient, err := age.NewScryptRecipient(passphrase)
		assert.NoError(t, err)
		recipient.SetWorkFactor(10)
		enc, err := age.Encrypt(w, recipient)
		assert.NoError(t, err)
		w = enc
		closers = append([]io.Closer{enc}, closers...)
	}
	switch Layers(name).Compression {
	case base.FileCompressionFormatGzip:
		gz := gzip.NewWriter(w)
		w = gz
		closers = append([]io.Closer{gz}, closers...)
	case base.FileCompressionFormatZstd:
		zw, err := zstd.NewWriter(w)
		assert.NoError(t, err)
		w = zw
		closers = append([]io.Closer{zw}, closers...)
	case base.FileCompressionNone, base.FileCompressionFormatZip, base.FileCompressionFormatTar:
	}
	_, err := w.Write(content)
	assert.NoError(t, err)
	for _, c := range closers {
		assert.NoError(t, c.Close())
	}
	return out.Bytes()
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

// A file a job saved is read back as the job's command printed it: decrypted
// as its .age says, decompressed as its .gz or .zst says.
func TestNewReader(t *testing.T) {
	content := []byte("CREATE TABLE notes(v text);\nINSERT INTO notes VALUES ('kept');\n")
	for _, name := range []string{"dump.sql", "dump.sql.gz", "dump.sql.zst", "dump.sql.age", "dump.sql.gz.age",
		"dump.sql.zst.age"} {
		t.Run(name, func(t *testing.T) {
			r, err := NewReader(bytes.NewReader(written(t, name, content)), name, passphrase)
			assert.NoError(t, err)
			got, err := io.ReadAll(r)
			assert.NoError(t, err)
			assert.Equal(t, content, got)
		})
	}
}

// An encrypted file is not read without its passphrase, nor with another.
func TestNewReaderRefusesAnEncryptedFileWithoutItsPassphrase(t *testing.T) {
	data := written(t, "dump.sql.gz.age", []byte("secret rows"))

	_, err := NewReader(bytes.NewReader(data), "dump.sql.gz.age", "")
	assert.True(t, errors.Is(err, hperrors.ErrMissing), "no passphrase: %v", err)

	_, err = NewReader(bytes.NewReader(data), "dump.sql.gz.age", "not the passphrase")
	assert.Error(t, err, "another passphrase")
}

// The layers are the name's, from the outside in: what is read first is last.
func TestLayers(t *testing.T) {
	assert.Equal(t, FileLayers{}, Layers("dump.sql"))
	assert.Equal(t, FileLayers{Compression: "gzip"}, Layers("dump.sql.gz"))
	assert.Equal(t, FileLayers{Compression: "zstd", Encrypted: true}, Layers("dump.sql.zst.age"))
	assert.Equal(t, FileLayers{Encrypted: true}, Layers("dump.age"))
	assert.Equal(t, FileLayers{}, Layers("dump.gz.sql"), "a suffix not last is the name's")
}
