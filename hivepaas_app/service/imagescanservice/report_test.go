package imagescanservice

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

var scannedAt = time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// An OS package's finding: the image, its OS past support, and the DB it was
// matched against.
func TestParseReportOfAnOSPackage(t *testing.T) {
	r, err := ParseReport(readFixture(t, "alpine-3.10.json"), "alpine:3.10", scannedAt)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "sha256:19a85ffa0ee14b7c64ca5104dc4e1c5021fa04bd6b4a692074691ce154f5d074", r.Scan.ImageID)
	assert.Equal(t, "alpine:3.10", r.Scan.ImageRef)
	assert.Equal(t, "linux/arm64", r.Scan.Platform)
	assert.Equal(t, "alpine", r.Scan.OSFamily)
	assert.Equal(t, "3.10.9", r.Scan.OSName)
	assert.True(t, r.Scan.OSEOSL)
	assert.Equal(t, "0.75.0", r.Scan.TrivyVersion)
	assert.False(t, r.Scan.DBUpdatedAt.IsZero())
	assert.Equal(t, base.ImageScanStatusDone, r.Scan.Status)

	if !assert.Len(t, r.Findings, 1) {
		return
	}
	f := r.Findings[0]
	assert.Equal(t, "CVE-2021-36159", f.VulnID)
	assert.Equal(t, "apk-tools", f.PkgName)
	assert.Equal(t, "alpine", f.PkgType)
	assert.Equal(t, "2.10.6-r0", f.InstalledVersion)
	assert.Equal(t, "2.10.7-r0", f.FixedVersion)
	assert.Equal(t, base.VulnSeverityCritical, f.Severity)
	assert.Equal(t, "https://avd.aquasec.com/nvd/cve-2021-36159", f.URL)
	assert.Equal(t, r.Scan.ImageID, f.ImageID)
}

// A jar's findings carry where the jar is; an image asked for by its ID is
// shown by its tag.
func TestParseReportOfAJar(t *testing.T) {
	r, err := ParseReport(readFixture(t, "log4j-jar.json"),
		"sha256:b2bdf71a218e2b1d9a8fbdb2b5a3c8fd109a3469dc3bafd4866ec339ad9f1007", scannedAt)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hp-scan-jar:test", r.Scan.ImageRef)
	if !assert.Len(t, r.Findings, 7) {
		return
	}
	for _, f := range r.Findings {
		assert.Equal(t, "jar", f.PkgType)
		assert.Equal(t, "app/log4j-core-2.14.1.jar", f.PkgPath)
		assert.Equal(t, "org.apache.logging.log4j:log4j-core", f.PkgName)
		assert.True(t, f.Fixable())
	}
	summary := Summarize(r.Findings, nil, scannedAt)
	assert.Equal(t, Count{Total: 2, Fixable: 2}, summary.Critical)
	assert.Equal(t, Count{Total: 1, Fixable: 1}, summary.High)
	assert.Equal(t, Count{Total: 4, Fixable: 4}, summary.Medium)
}

// A finding reported twice for the same place is one; a severity Trivy may
// add later is UNKNOWN.
func TestParseReportKeepsOneOfEachFinding(t *testing.T) {
	data := []byte(`{"Trivy":{"Version":"0.75.0"},"Metadata":{"ImageID":"sha256:a"},"Results":[
		{"Type":"debian","Vulnerabilities":[
			{"VulnerabilityID":"CVE-1","PkgName":"libc6","InstalledVersion":"1","Severity":"HIGH"},
			{"VulnerabilityID":"CVE-1","PkgName":"libc6","InstalledVersion":"1","Severity":"HIGH"},
			{"VulnerabilityID":"CVE-2","PkgName":"libc6","InstalledVersion":"1","Severity":"NEGLIGIBLE"}]}]}`)
	r, err := ParseReport(data, "", scannedAt)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, r.Findings, 2) {
		return
	}
	assert.Equal(t, base.VulnSeverityUnknown, r.Findings[1].Severity)
	assert.Equal(t, "sha256:a", r.Scan.ImageID)
	assert.Empty(t, r.Scan.Platform)
}

// A report that names no image is no scan.
func TestParseReportWithoutAnImage(t *testing.T) {
	_, err := ParseReport([]byte(`{"Results":[]}`), "x", scannedAt)
	assert.Error(t, err)
	_, err = ParseReport([]byte(`not json`), "x", scannedAt)
	assert.Error(t, err)
}
