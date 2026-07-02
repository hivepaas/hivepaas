package httputil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func TestParseRequestLang(t *testing.T) {
	// Setup available languages in the translation package via Mock/Assume en is available
	assert.Equal(t, translation.Lang("en"), ParseRequestLang("en-US,en;q=0.9"))
	assert.Equal(t, translation.Lang("en"), ParseRequestLang("en"))
	assert.Equal(t, translation.Lang("en"), ParseRequestLang("en-GB"))

	// Default fallback
	assert.Equal(t, translation.GetDefaultLang(), ParseRequestLang("invalid_lang_code"))
	assert.Equal(t, translation.GetDefaultLang(), ParseRequestLang(""))
}

func TestRenderMultipartForm(t *testing.T) {
	// Test text fields
	fields := []*MultipartFormField{
		{Name: "field1", Value: "value1"},
		{Name: "field2", Value: "value2"},
	}

	r, contentType, err := RenderMultipartForm(fields)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(contentType, "multipart/form-data"))

	data, err := io.ReadAll(r)
	assert.NoError(t, err)
	strData := string(data)
	assert.Contains(t, strData, `name="field1"`)
	assert.Contains(t, strData, "value1")
	assert.Contains(t, strData, `name="field2"`)
	assert.Contains(t, strData, "value2")

	// Test file fields
	fieldsWithFile := []*MultipartFormField{
		{Name: "file", Value: "test.txt", FileData: bytes.NewBufferString("file content")},
	}
	r, _, err = RenderMultipartForm(fieldsWithFile)
	assert.NoError(t, err)

	data, err = io.ReadAll(r)
	assert.NoError(t, err)
	strData = string(data)
	assert.Contains(t, strData, `name="file"`)
	assert.Contains(t, strData, `filename="test.txt"`)
	assert.Contains(t, strData, "file content")
}

func TestHTTPGet(t *testing.T) {
	// Test successful GET
	tsSuccess := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success response"))
	}))
	defer tsSuccess.Close()

	data, err := HTTPGet(context.Background(), tsSuccess.URL)
	assert.NoError(t, err)
	assert.Equal(t, "success response", string(data))

	// Test GET with error status code
	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer tsError.Close()

	data, err = HTTPGet(context.Background(), tsError.URL)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, ErrHTTPStatus))
	assert.Contains(t, err.Error(), "http status error: 404")
}

func TestHTTPPost(t *testing.T) {
	// Test successful POST
	tsSuccess := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "custom-header-value", r.Header.Get("X-Custom-Header"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "post body", string(body))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success response"))
	}))
	defer tsSuccess.Close()

	data, err := HTTPPost(
		context.Background(),
		tsSuccess.URL,
		bytes.NewBufferString("post body"),
		func(r *http.Request) {
			r.Header.Set("X-Custom-Header", "custom-header-value")
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "success response", string(data))

	// Test POST with error status code
	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer tsError.Close()

	data, err = HTTPPost(context.Background(), tsError.URL, nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.True(t, errors.Is(err, ErrHTTPStatus))
	assert.Contains(t, err.Error(), "http status error: 500")
}
