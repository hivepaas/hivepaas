package sslcertuc

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ZipEntries(t *testing.T) {
	certContent := "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\n-----END CERTIFICATE-----"
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA0Y3...\n-----END RSA PRIVATE KEY-----"
	caContent := "-----BEGIN CERTIFICATE-----\nCA_CERT_CONTENT\n-----END CERTIFICATE-----"

	// Case 1: Without CA cert
	t.Run("without CA cert", func(t *testing.T) {
		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		certWriter, err := zipWriter.Create(certEntryName)
		assert.NoError(t, err)
		_, err = certWriter.Write([]byte(certContent))
		assert.NoError(t, err)

		keyWriter, err := zipWriter.Create(keyEntryName)
		assert.NoError(t, err)
		_, err = keyWriter.Write([]byte(keyContent))
		assert.NoError(t, err)

		assert.NoError(t, zipWriter.Close())

		zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		assert.NoError(t, err)
		assert.Len(t, zipReader.File, 2)

		filesMap := make(map[string]string)
		for _, f := range zipReader.File {
			rc, err := f.Open()
			assert.NoError(t, err)
			content, err := io.ReadAll(rc)
			assert.NoError(t, err)
			assert.NoError(t, rc.Close())
			filesMap[f.Name] = string(content)
		}

		assert.Equal(t, certContent, filesMap[certEntryName])
		assert.Equal(t, keyContent, filesMap[keyEntryName])
		assert.NotContains(t, filesMap, caCertEntryName)
	})

	// Case 2: With CA cert
	t.Run("with CA cert", func(t *testing.T) {
		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		certWriter, err := zipWriter.Create(certEntryName)
		assert.NoError(t, err)
		_, err = certWriter.Write([]byte(certContent))
		assert.NoError(t, err)

		keyWriter, err := zipWriter.Create(keyEntryName)
		assert.NoError(t, err)
		_, err = keyWriter.Write([]byte(keyContent))
		assert.NoError(t, err)

		caWriter, err := zipWriter.Create(caCertEntryName)
		assert.NoError(t, err)
		_, err = caWriter.Write([]byte(caContent))
		assert.NoError(t, err)

		assert.NoError(t, zipWriter.Close())

		zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		assert.NoError(t, err)
		assert.Len(t, zipReader.File, 3)

		filesMap := make(map[string]string)
		for _, f := range zipReader.File {
			rc, err := f.Open()
			assert.NoError(t, err)
			content, err := io.ReadAll(rc)
			assert.NoError(t, err)
			assert.NoError(t, rc.Close())
			filesMap[f.Name] = string(content)
		}

		assert.Equal(t, certContent, filesMap[certEntryName])
		assert.Equal(t, keyContent, filesMap[keyEntryName])
		assert.Equal(t, caContent, filesMap[caCertEntryName])
	})
}
