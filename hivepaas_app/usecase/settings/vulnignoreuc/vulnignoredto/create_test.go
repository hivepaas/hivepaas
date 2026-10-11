package vulnignoredto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createReq(vulnID, pkg, reason string, expireAt *time.Time) *CreateVulnIgnoreReq {
	req := NewCreateVulnIgnoreReq()
	req.VulnIgnoreBaseReq = &VulnIgnoreBaseReq{VulnID: vulnID, Package: pkg, Reason: reason, ExpireAt: expireAt}
	_ = req.ModifyRequest()
	return req
}

// An ignore names a vulnerability as Trivy does, says why, and expires within
// a year when it expires.
func TestCreateVulnIgnoreValidation(t *testing.T) {
	soon := time.Now().Add(30 * 24 * time.Hour)
	req := createReq(" cve-2021-36159 ", " apk-tools ", " the fetcher is never used ", &soon)
	assert.Empty(t, req.Validate())
	assert.Equal(t, "CVE-2021-36159", req.VulnID)
	assert.Equal(t, "apk-tools", req.Package)
	assert.Equal(t, "the fetcher is never used", req.Reason)
	assert.Equal(t, "CVE-2021-36159 (apk-tools)", req.ToEntity(nil).NameOf())
	assert.Empty(t, createReq("GHSA-p6xc-xr62-6r2g", "", "r", nil).Validate())
	assert.Equal(t, "GHSA-P6XC-XR62-6R2G", createReq("GHSA-p6xc-xr62-6r2g", "", "r", nil).ToEntity(nil).NameOf())

	assert.NotEmpty(t, createReq("CVE 2021", "", "r", nil).Validate(), "not an ID")
	assert.NotEmpty(t, createReq("CVE-2021-1", "", "  ", nil).Validate(), "no reason")
	past := time.Now().Add(-time.Hour)
	assert.NotEmpty(t, createReq("CVE-2021-1", "", "r", &past).Validate(), "expired already")
	far := time.Now().Add(2 * 365 * 24 * time.Hour)
	assert.NotEmpty(t, createReq("CVE-2021-1", "", "r", &far).Validate(), "more than a year away")
	assert.True(t, createReq("CVE-2021-1", "", "r", &time.Time{}).ExpireAtOrZero().IsZero())
}
