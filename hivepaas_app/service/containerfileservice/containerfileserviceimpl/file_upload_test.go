package containerfileserviceimpl

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
)

func TestPrepareUploadTarStream_SingleFile(t *testing.T) {
	svc := &service{}
	fileContent := []byte("hello world")

	req := &containerfileservice.PrepareUploadTarStreamReq{
		Path:     "/app/config.json",
		FileName: "config.json",
		FileSize: int64(len(fileContent)),
		Extract:  false,
		Content:  io.NopCloser(bytes.NewReader(fileContent)),
	}

	resp, err := svc.PrepareUploadTarStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected resp not to be nil")
	}
	if resp.DestPath != "/app" {
		t.Errorf("expected DstDir to be '/app', got '%s'", resp.DestPath)
	}

	// Read generated tar stream
	tarReader := tar.NewReader(resp.TarStream)
	header, err := tarReader.Next()
	if err != nil {
		t.Fatalf("unexpected header error: %v", err)
	}
	if header.Name != "config.json" {
		t.Errorf("expected header name 'config.json', got '%s'", header.Name)
	}
	if header.Size != int64(len(fileContent)) {
		t.Errorf("expected header size %d, got %d", len(fileContent), header.Size)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, tarReader)
	if err != nil {
		t.Fatalf("unexpected copy error: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), fileContent) {
		t.Errorf("expected '%s', got '%s'", fileContent, buf.String())
	}

	_, err = tarReader.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	_ = resp.TarStream.Close()
}

func TestPrepareUploadTarStream_ExtractTarGz_WithTarSlip(t *testing.T) {
	svc := &service{}

	// Create a tar.gz with a normal file and a tar slip attempt (../../etc/passwd)
	var gzBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&gzBuf)
	tarWriter := tar.NewWriter(gzWriter)

	// File 1: safe
	file1Content := []byte("index html content")
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "public/index.html",
		Size: int64(len(file1Content)),
		Mode: 0644,
	}); err != nil {
		t.Fatalf("write header error: %v", err)
	}
	_, _ = tarWriter.Write(file1Content)

	// File 2: tar slip attempt
	file2Content := []byte("malicious")
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "../../../../etc/passwd",
		Size: int64(len(file2Content)),
		Mode: 0644,
	}); err != nil {
		t.Fatalf("write header error: %v", err)
	}
	_, _ = tarWriter.Write(file2Content)

	_ = tarWriter.Close()
	_ = gzWriter.Close()

	req := &containerfileservice.PrepareUploadTarStreamReq{
		Path:              "/var/www/html",
		FileName:          "bundle.tar.gz",
		FileSize:          int64(gzBuf.Len()),
		Extract:           true,
		CompressionFormat: base.FileCompressionFormatGzip,
		Content:           io.NopCloser(&gzBuf),
	}

	resp, err := svc.PrepareUploadTarStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DestPath != "/var/www/html" {
		t.Errorf("expected DstDir '/var/www/html', got '%s'", resp.DestPath)
	}

	// Inspect sanitized tar stream
	tarReader := tar.NewReader(resp.TarStream)

	// Entry 1
	h1, err := tarReader.Next()
	if err != nil {
		t.Fatalf("read entry 1 error: %v", err)
	}
	if h1.Name != "public/index.html" {
		t.Errorf("expected 'public/index.html', got '%s'", h1.Name)
	}

	// Entry 2 (must be sanitized: etc/passwd without ../)
	h2, err := tarReader.Next()
	if err != nil {
		t.Fatalf("read entry 2 error: %v", err)
	}
	if h2.Name != "etc/passwd" {
		t.Errorf("expected 'etc/passwd', got '%s'", h2.Name)
	}

	_, err = tarReader.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	_ = resp.TarStream.Close()
}

func TestPrepareUploadTarStream_ExtractZip(t *testing.T) {
	svc := &service{}

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	f, err := zipWriter.Create("styles/main.css")
	if err != nil {
		t.Fatalf("zip create error: %v", err)
	}
	_, err = f.Write([]byte("body { color: red; }"))
	if err != nil {
		t.Fatalf("zip write error: %v", err)
	}
	_ = zipWriter.Close()

	req := &containerfileservice.PrepareUploadTarStreamReq{
		Path:     "/var/www/html",
		FileName: "assets.zip",
		FileSize: int64(zipBuf.Len()),
		Extract:  true,
		Content:  io.NopCloser(&zipBuf),
	}

	resp, err := svc.PrepareUploadTarStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DestPath != "/var/www/html" {
		t.Errorf("expected DstDir '/var/www/html', got '%s'", resp.DestPath)
	}

	tarReader := tar.NewReader(resp.TarStream)
	h, err := tarReader.Next()
	if err != nil {
		t.Fatalf("tar read error: %v", err)
	}
	if h.Name != "styles/main.css" {
		t.Errorf("expected 'styles/main.css', got '%s'", h.Name)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, tarReader)
	if err != nil {
		t.Fatalf("tar copy error: %v", err)
	}
	if buf.String() != "body { color: red; }" {
		t.Errorf("expected 'body { color: red; }', got '%s'", buf.String())
	}
	_ = resp.TarStream.Close()
}
