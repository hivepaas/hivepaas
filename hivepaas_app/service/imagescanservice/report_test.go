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
	r, err := ParseReport(readFixture(t, "alpine-3.10.json"), "", "alpine:3.10", scannedAt)
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
	r, err := ParseReport(readFixture(t, "log4j-jar.json"), "",
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
	r, err := ParseReport(data, "", "", scannedAt)
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
	_, err := ParseReport([]byte(`{"Results":[]}`), "", "x", scannedAt)
	assert.Error(t, err)
	_, err = ParseReport([]byte(`not json`), "", "x", scannedAt)
	assert.Error(t, err)
}

// Read from its registry, an image's report names its config's digest: the
// scan is kept by the ID the node running it gives it.
func TestParseReportKeepsTheNodesImageID(t *testing.T) {
	data := []byte(`{"Metadata":{"ImageID":"sha256:config"},"Results":[]}`)
	r, err := ParseReport(data, "sha256:index", "alpine:3.10", scannedAt)
	if assert.NoError(t, err) {
		assert.Equal(t, "sha256:index", r.Scan.ImageID)
	}
}

// What a scan's container prints: the exit code, the report, and the DB's
// date from trivy version, which a standalone report lacks.
func TestParseScanOutput(t *testing.T) {
	out, err := ParseScanOutput([]byte(`{"exitCode":0,"version":{"Version":"0.75.0","VulnerabilityDB":` +
		`{"UpdatedAt":"2026-10-14T07:11:52Z"}},"report":{"Metadata":{"ImageID":"sha256:a"}}}`))
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, out.HasReport())
	assert.Equal(t, time.Date(2026, 10, 14, 7, 11, 52, 0, time.UTC), out.DBUpdatedAt())

	out, err = ParseScanOutput([]byte(`{"exitCode":1,"version":null,"report":null}`))
	if assert.NoError(t, err) {
		assert.False(t, out.HasReport())
		assert.True(t, out.DBUpdatedAt().IsZero())
	}
}

// A Go binary's findings, from a scan of it as a root filesystem - WAL-G's: by
// module, with no image or OS of its own, kept by the ID it was asked by. The
// binary's own module, built from a commit, is matched by its pseudo-version.
func TestParseReportOfAGoBinary(t *testing.T) {
	const sha = "sha256:965a7147852435bc4b54f272799b63df826b54ef65cf7c4fc688c82b96c0c4eb"
	r, err := ParseReport(readFixture(t, "walg-gobinary.json"), sha, "wal-g v3.0.8", scannedAt)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, sha, r.Scan.ImageID)
	assert.Equal(t, "wal-g v3.0.8", r.Scan.ImageRef)
	assert.Empty(t, r.Scan.Platform, "the report names none: the caller knows the build's")
	assert.Empty(t, r.Scan.OSFamily)
	assert.Equal(t, "0.75.0", r.Scan.TrivyVersion)
	if !assert.Len(t, r.Findings, 4) {
		return
	}
	byID := map[string]string{}
	for _, f := range r.Findings {
		assert.Equal(t, "gobinary", f.PkgType)
		assert.Equal(t, sha, f.ImageID)
		byID[f.VulnID] = f.PkgName + " " + f.InstalledVersion + " " + string(f.Severity) + " " + f.FixedVersion
	}
	assert.Equal(t, "golang.org/x/net v0.58.0 HIGH 0.60.0", byID["CVE-2026-78669"])
	assert.Equal(t, "golang.org/x/crypto v0.55.0 UNKNOWN ", byID["GO-2026-5932"])
	assert.Equal(t, "github.com/wal-g/wal-g v0.0.0-20260820091616-3e493188db28+dirty HIGH 1.1",
		byID["CVE-2021-38599"])
}
