package httputil

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

type MultipartFormField struct {
	Name     string
	Value    string
	FileData io.Reader
}

// RenderMultipartForm creates data representation of multipart form
func RenderMultipartForm(values []*MultipartFormField) (io.Reader, string, error) {
	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)

	for _, field := range values {
		if field.FileData == nil {
			err := mpw.WriteField(field.Name, field.Value)
			if err != nil {
				return nil, "", tracerr.Wrap(err)
			}
		} else {
			fieldWriter, err := mpw.CreateFormFile(field.Name, field.Value)
			if err != nil {
				return nil, "", tracerr.Wrap(err)
			}
			_, err = io.Copy(fieldWriter, field.FileData)
			if err != nil {
				return nil, "", tracerr.Wrap(err)
			}
		}
	}

	// Close the multipart writer before creating the request
	if err := mpw.Close(); err != nil {
		return nil, "", tracerr.Wrap(err)
	}
	return buf, mpw.FormDataContentType(), nil
}
