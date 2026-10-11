package vulnignoredto

import (
	"regexp"
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	vulnIDMaxLen  = 100
	packageMaxLen = 255
	reasonMaxLen  = 1000

	// maxExpireIn is how far away an ignore may expire: a year, so each is
	// looked at again - and a day, for the clock of the browser that asked.
	maxExpireIn = timeutil.Dur365Days + timeutil.Day
)

// vulnIDPattern is a vulnerability's ID as Trivy reports it: CVE-2025-1234,
// GHSA-xxxx-xxxx-xxxx, RUSTSEC-2024-0001.
var vulnIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*(-[A-Za-z0-9.:_]+)+$`)

type CreateVulnIgnoreReq struct {
	settings.CreateSettingReq
	*VulnIgnoreBaseReq
}

type VulnIgnoreBaseReq struct {
	VulnID  string `json:"vulnId"`
	Package string `json:"package"`
	Reason  string `json:"reason"`
	// ExpireAt is when the ignore stops applying; null for never, and at most a
	// year away.
	ExpireAt *time.Time `json:"expireAt"`
}

func (req *VulnIgnoreBaseReq) ToEntity(auth *basedto.Auth) *entity.VulnerabilityIgnore {
	ignore := &entity.VulnerabilityIgnore{VulnID: req.VulnID, Package: req.Package, Reason: req.Reason}
	if auth != nil && auth.User != nil && auth.User.User != nil {
		ignore.SetBy = auth.User.Username
	}
	return ignore
}

// ExpireAtOrZero is the expiry the setting is given: zero for none.
func (req *VulnIgnoreBaseReq) ExpireAtOrZero() time.Time {
	if req.ExpireAt == nil {
		return time.Time{}
	}
	return req.ExpireAt.UTC()
}

func (req *VulnIgnoreBaseReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.VulnID = strings.ToUpper(strings.TrimSpace(req.VulnID))
	req.Package = strings.TrimSpace(req.Package)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.ExpireAt != nil && req.ExpireAt.IsZero() {
		req.ExpireAt = nil
	}
	return nil
}

func (req *VulnIgnoreBaseReq) validate() (res []vld.Validator) {
	if req == nil {
		return nil
	}
	res = append(res, basedto.ValidateStr(&req.VulnID, true, 1, vulnIDMaxLen, "vulnId")...)
	res = append(res, vld.StrMatch(&req.VulnID, vulnIDPattern).OnError(vld.SetField("vulnId", nil)))
	res = append(res, basedto.ValidateStr(&req.Package, false, 1, packageMaxLen, "package")...)
	res = append(res, basedto.ValidateStr(&req.Reason, true, 1, reasonMaxLen, "reason")...)
	if req.ExpireAt != nil {
		now := time.Now()
		res = append(res, basedto.ValidateTime(req.ExpireAt, true, now, now.Add(maxExpireIn), "expireAt")...)
	}
	return res
}

func NewCreateVulnIgnoreReq() *CreateVulnIgnoreReq {
	return &CreateVulnIgnoreReq{}
}

func (req *CreateVulnIgnoreReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateVulnIgnoreReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateVulnIgnoreResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
